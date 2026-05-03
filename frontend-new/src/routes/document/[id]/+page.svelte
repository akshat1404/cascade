<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { page } from "$app/stores";
	import { Editor } from "@tiptap/core";
	import StarterKit from "@tiptap/starter-kit";
	import { TextStyle } from "@tiptap/extension-text-style";
	import { FontFamily } from "@tiptap/extension-font-family";
	import { Color } from "@tiptap/extension-color";
	import { Highlight } from "@tiptap/extension-highlight";
	import { TextAlign } from "@tiptap/extension-text-align";
	import { Underline } from "@tiptap/extension-underline";
	import { Table } from "@tiptap/extension-table";
	import { TableRow } from "@tiptap/extension-table-row";
	import { TableCell } from "@tiptap/extension-table-cell";
	import { TableHeader } from "@tiptap/extension-table-header";
	import { Image } from "@tiptap/extension-image";
	import { NodeSelection } from "@tiptap/pm/state";
	import { marked } from "marked";
	import { supabase } from "$lib/supabase";
	import imageCompression from "browser-image-compression";
	import { saveOfflineDoc, getOfflineDoc, deleteOfflineDoc } from "$lib/offlineDb";

	// ── Custom Image extension with style + width attributes ──────
	const CustomImage = Image.extend({
		addAttributes() {
			return {
				...this.parent?.(),
				width: {
					default: null,
					parseHTML: (el: HTMLElement) => el.getAttribute('width'),
					renderHTML: (attrs: Record<string, unknown>) =>
						attrs.width ? { width: attrs.width as string } : {},
				},
				style: {
					default: null,
					parseHTML: (el: HTMLElement) => el.getAttribute('style'),
					renderHTML: (attrs: Record<string, unknown>) =>
						attrs.style ? { style: attrs.style as string } : {},
				},
			};
		},
	});

	let { data } = $props();

	const docId = $page.params.id;

	let editor = $state<Editor | null>(null);
	let editorEl: HTMLDivElement;
	let title = $state<string>(data.document.title ?? "");
	let saveStatus = $state<"Saved" | "Saving..." | "Unsaved" | "Offline">("Saved");
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	// ── Offline handling ──────────────────────────────────────
	let isOnline = $state(typeof navigator !== 'undefined' ? navigator.onLine : true);
	let hasPendingOfflineSave = $state(false);

	let isBold = $derived(editor?.isActive("bold") ?? false);
	let isItalic = $derived(editor?.isActive("italic") ?? false);
	let isUnderline = $derived(editor?.isActive("underline") ?? false);
	let isStrike = $derived(editor?.isActive("strike") ?? false);
	let isH1 = $derived(editor?.isActive("heading", { level: 1 }) ?? false);
	let isH2 = $derived(editor?.isActive("heading", { level: 2 }) ?? false);
	let isH3 = $derived(editor?.isActive("heading", { level: 3 }) ?? false);
	let isBulletList = $derived(editor?.isActive("bulletList") ?? false);
	let isOrderedList = $derived(editor?.isActive("orderedList") ?? false);
	let isAlignLeft = $derived(
		editor?.isActive({ textAlign: "left" }) ?? false,
	);
	let isAlignCenter = $derived(
		editor?.isActive({ textAlign: "center" }) ?? false,
	);
	let isAlignRight = $derived(
		editor?.isActive({ textAlign: "right" }) ?? false,
	);
	let isAlignJustify = $derived(
		editor?.isActive({ textAlign: "justify" }) ?? false,
	);

	let currentColor = $state("#000000");
	let currentFontSize = $state("16px");

	// ── Bubble menu ──────────────────────────────
	let showBubble = $state(false);
	let bubbleX = $state(0);
	let bubbleY = $state(0);
	let bubbleInner = $state<HTMLDivElement | null>(null);

	// AI bubble state
	let aiLoading = $state<string | null>(null);            // which action is loading
	let aiCustomPrompt = $state("");                        // free-text Ask AI prompt
	let savedSelection = $state<{ from: number; to: number } | null>(null); // locked selection
	let showAskModal = $state(false);                       // Ask AI modal visibility

	// ── Image upload state ─────────────────────────────────────
	let imageFileInput = $state<HTMLInputElement | null>(null);
	let imageUploading = $state(false);
	let imageUploadError = $state<string | null>(null);
	let imageUploadToast = $state<"idle" | "compressing" | "uploading" | "done" | "error">("idle");

	// ── Image selection + editing state ────────────────────────
	let isImageSelected = $state(false);
	let selectedImgAlt = $state("");
	let imgReplaceInput = $state<HTMLInputElement | null>(null);

	// Filter / style sliders (populated from selected image attrs)
	let imgWidthPct   = $state(100);   // 10–100 %
	let imgBrightness = $state(100);   // 10–200 %
	let imgContrast   = $state(100);   // 10–200 %
	let imgSaturation = $state(100);   // 0–200  %
	let imgBlur       = $state(0);     // 0–20   px
	let imgGrayscale  = $state(0);     // 0–100  %
	let imgSepia      = $state(0);     // 0–100  %
	let imgOpacity    = $state(100);   // 10–100 %
	let imgRadius     = $state(0);     // 0–50   px
	let imgRotate     = $state(0);     // -180–180 deg
	// Positional translate (drag)
	let imgTX = $state(0);  // translate X px
	let imgTY = $state(0);  // translate Y px

	// Drag tracking (non-reactive, no $state needed)
	let _dragging = false;
	let _dragSX = 0, _dragSY = 0;  // mouse start
	let _dragSTX = 0, _dragSTY = 0; // translate start

	// ── Parse/build image style ────────────────────────────────
	function parseImageStyle(style: string) {
		const num = (re: RegExp, def: number) => { const m = style.match(re); return m ? parseFloat(m[1]) : def; };
		imgBrightness = num(/brightness\((\d+(?:\.\d+)?)%\)/, 100);
		imgContrast   = num(/contrast\((\d+(?:\.\d+)?)%\)/, 100);
		imgSaturation = num(/saturate\((\d+(?:\.\d+)?)%\)/, 100);
		imgBlur       = num(/blur\((\d+(?:\.\d+)?)px\)/, 0);
		imgGrayscale  = num(/grayscale\((\d+(?:\.\d+)?)%\)/, 0);
		imgSepia      = num(/sepia\((\d+(?:\.\d+)?)%\)/, 0);
		imgOpacity    = num(/opacity:\s*(\d+(?:\.\d+)?)%/, 100);
		imgRadius     = num(/border-radius:\s*(\d+(?:\.\d+)?)px/, 0);
		imgRotate     = num(/rotate\((-?\d+(?:\.\d+)?)deg\)/, 0);
		const txm = style.match(/translateX\((-?\d+(?:\.\d+)?)px\)/);
		const tym = style.match(/translateY\((-?\d+(?:\.\d+)?)px\)/);
		imgTX = txm ? parseFloat(txm[1]) : 0;
		imgTY = tym ? parseFloat(tym[1]) : 0;
		const wm = style.match(/width:\s*(\d+(?:\.\d+)?)%/);
		imgWidthPct = wm ? parseFloat(wm[1]) : 100;
	}

	function buildImageStyle(): string {
		const f = `brightness(${imgBrightness}%) contrast(${imgContrast}%) saturate(${imgSaturation}%) blur(${imgBlur}px) grayscale(${imgGrayscale}%) sepia(${imgSepia}%)`;
		return [
			`width: ${imgWidthPct}%`,
			`filter: ${f}`,
			`opacity: ${imgOpacity}%`,
			`border-radius: ${imgRadius}px`,
			`transform: translateX(${imgTX}px) translateY(${imgTY}px) rotate(${imgRotate}deg)`,
			`display: block`,
		].join('; ');
	}

	// Helper: get the currently-selected image DOM element
	function getSelectedImgEl(): HTMLImageElement | null {
		if (!editor) return null;
		const { selection } = editor.state;
		if (!(selection instanceof NodeSelection) || selection.node.type.name !== 'image') return null;
		const dom = editor.view.nodeDOM(selection.from) as HTMLElement | null;
		if (!dom) return null;
		return (dom.tagName === 'IMG' ? dom : dom.querySelector('img')) as HTMLImageElement | null;
	}

	function applyImageStyle() {
		if (!editor) return;
		const style = buildImageStyle();
		// 1. Apply immediately to DOM for instant visual feedback
		const imgEl = getSelectedImgEl();
		if (imgEl) imgEl.setAttribute('style', style);
		// 2. Commit to ProseMirror state so it saves
		editor.chain().updateAttributes('image', { style }).run();
	}

	function updateBubble(e: Editor) {
		const { selection } = e.state;

		// ── Image node? Use proper NodeSelection instanceof check
		if (selection instanceof NodeSelection && selection.node.type.name === 'image') {
			const node = selection.node;
			selectedImgAlt = node.attrs.alt ?? '';
			parseImageStyle(node.attrs.style ?? '');
			isImageSelected = true;
			showBubble = false;
			if (!showAskModal) savedSelection = null;
			return;
		}

		// ── Not an image — reset
		isImageSelected = false;

		// ── Normal text selection
		const { from, to } = selection;
		if (from === to) {
			showBubble = false;
			if (!showAskModal) savedSelection = null;
			return;
		}
		const sel = window.getSelection();
		if (!sel || sel.rangeCount === 0 || sel.isCollapsed) { showBubble = false; savedSelection = null; return; }
		const rect = sel.getRangeAt(0).getBoundingClientRect();
		if (!rect.width) { showBubble = false; savedSelection = null; return; }
		const halfW = 160;
		bubbleX = Math.min(Math.max(rect.left + rect.width / 2, halfW), window.innerWidth - halfW);
		bubbleY = rect.top - 10;
		savedSelection = { from, to };
		showBubble = true;
		
	}

	// ── Drag-to-position ────────────────────────────────────────
	function onEditorMouseDown(e: MouseEvent) {
		const target = e.target as HTMLElement;
		// Only drag if target is an image that is already ProseMirror-selected
		// (the selection class is set on the *first* click; drag starts on the *second* hold)
		if (target.tagName !== 'IMG') return;
		if (!target.classList.contains('ProseMirror-selectednode')) return;
		e.preventDefault();
		_dragging = true;
		_dragSX = e.clientX; _dragSY = e.clientY;
		_dragSTX = imgTX;    _dragSTY = imgTY;
		window.addEventListener('mousemove', onWindowMouseMove);
		window.addEventListener('mouseup',   onWindowMouseUp);
	}
	function onWindowMouseMove(e: MouseEvent) {
		if (!_dragging) return;
		imgTX = _dragSTX + (e.clientX - _dragSX);
		imgTY = _dragSTY + (e.clientY - _dragSY);
		// Live-update style without going through ProseMirror (for smooth drag)
		if (!editor) return;
		const { selection } = editor.state;
		if (!(selection instanceof NodeSelection)) return;
		const domEl = editor.view.nodeDOM(selection.from) as HTMLElement | null;
		const img = domEl?.tagName === 'IMG' ? domEl : domEl?.querySelector('img');
		if (img) (img as HTMLElement).style.transform =
			`translateX(${imgTX}px) translateY(${imgTY}px) rotate(${imgRotate}deg)`;
	}
	function onWindowMouseUp() {
		if (!_dragging) return;
		_dragging = false;
		window.removeEventListener('mousemove', onWindowMouseMove);
		window.removeEventListener('mouseup',   onWindowMouseUp);
		applyImageStyle(); // commit to ProseMirror
	}

	// ── Image actions ───────────────────────────────────────────
	function getImgPos(): number | null {
		if (!editor) return null;
		const sel = editor.state.selection;
		if (!(sel instanceof NodeSelection) || sel.node.type.name !== 'image') return null;
		return sel.from;
	}
	function deleteImage() {
		if (getImgPos() === null || !editor) return;
		editor.chain().focus().deleteSelection().run();
		isImageSelected = false;
	}
	function setImagePosition(float: 'left' | 'right' | 'center' | 'block') {
		if (!editor) return;
		if (float === 'center') {
			editor.chain().focus().updateAttributes('image', { style: buildImageStyle().replace(/display:[^;]+/, 'display:block; margin:0 auto') }).run();
		} else if (float === 'block') {
			applyImageStyle();
		} else {
			const margin = float === 'left' ? '0 1.5em 1em 0' : '0 0 1em 1.5em';
			editor.chain().focus().updateAttributes('image', { style: buildImageStyle() + `; float: ${float}; margin: ${margin}` }).run();
		}
	}

	const fontSizes = [
		"12px",
		"14px",
		"16px",
		"18px",
		"20px",
		"24px",
		"28px",
		"32px",
	];

	function scheduleSave() {
		saveStatus = "Unsaved";
		if (debounceTimer) clearTimeout(debounceTimer);
		if (!isOnline) {
			// Persist locally (async); will sync when back online
			void saveOffline();
			return;
		}
		debounceTimer = setTimeout(saveDocument, 1500);
	}

	/** Persist current editor state to IndexedDB for offline resilience. */
	async function saveOffline() {
		try {
			await saveOfflineDoc(docId, title, editor?.getJSON());
			hasPendingOfflineSave = true;
			saveStatus = "Offline";
		} catch (e) {
			console.warn('Could not write offline save to IndexedDB:', e);
		}
	}

	/** Flush the offline-buffered save to the server. Called when coming back online. */
	async function flushOfflineSave() {
		const pending = await getOfflineDoc(docId).catch(() => undefined);
		if (!pending) return;
		// Always use the most current editor state (not the stale snapshot)
		try {
			await saveDocument();
			await deleteOfflineDoc(docId);
			hasPendingOfflineSave = false;
		} catch {
			// Keep the IndexedDB record — retry on the next 'online' event
		}
	}

	function handleOnline() {
		isOnline = true;
		// Immediately sync any changes that accumulated while offline
		if (hasPendingOfflineSave || saveStatus === "Offline" || saveStatus === "Unsaved") {
			flushOfflineSave();
		}
	}

	function handleOffline() {
		isOnline = false;
		// Cancel in-flight debounce so the fetch attempt is abandoned
		if (debounceTimer) {
			clearTimeout(debounceTimer);
			debounceTimer = null;
		}
		// Capture whatever is currently in the editor
		saveOffline(); // async — fire-and-forget is fine here
	}

	async function saveDocument() {
		if (!isOnline) {
			await saveOffline();
			return;
		}
		saveStatus = "Saving...";
		try {
			const {
				data: { session },
			} = await supabase.auth.getSession();

			await fetch(`${import.meta.env.VITE_API_URL}/documents/${docId}`, {
				method: "PUT",
				headers: {
					"Content-Type": "application/json",
					Authorization: `Bearer ${session?.access_token}`,
				},
				body: JSON.stringify({
					title,
					content: editor?.getJSON(),
				}),
			});
			saveStatus = "Saved";
		} catch {
			saveStatus = "Unsaved";
		}
	}

	onMount(async () => {
		// Register online/offline listeners
		window.addEventListener('online',  handleOnline);
		window.addEventListener('offline', handleOffline);

		// If we already have an offline save from a previous session, mark it
		try {
			const pending = await getOfflineDoc(docId);
			if (pending) {
				hasPendingOfflineSave = true;
				if (isOnline) {
					// We're online — flush once the editor is fully mounted
					setTimeout(flushOfflineSave, 500);
				} else {
					saveStatus = "Offline";
				}
			}
		} catch (e) {
			console.warn('Could not read IndexedDB offline save:', e);
		}

		editor = new Editor({
			element: editorEl,
			extensions: [
				StarterKit,
				Underline,
				TextStyle,
				FontFamily,
				Color,
				Highlight.configure({ multicolor: true }),
				TextAlign.configure({ types: ["heading", "paragraph"] }),
				Table.configure({ resizable: false }),
				TableRow,
				TableCell,
				TableHeader,
				CustomImage.configure({ inline: false, allowBase64: false }),
			],
			content: data.document.content ?? "",
			onTransaction({ editor: e }) {
				editor = e;
				scheduleSave();
				updateBubble(e);
			},
			onSelectionUpdate({ editor: e }) {
				editor = e;
				updateBubble(e);
			},
		});
	});

	onDestroy(() => {
		if (debounceTimer) clearTimeout(debounceTimer);
		window.removeEventListener('online',  handleOnline);
		window.removeEventListener('offline', handleOffline);
		editor?.destroy();
	});

	function applyFontSize(size: string) {
		currentFontSize = size;
		editor?.chain().focus().setMark("textStyle", { fontSize: size }).run();
	}

	function applyColor(color: string) {
		currentColor = color;
		editor?.chain().focus().setColor(color).run();
	}

	let showExportMenu = $state(false);

	function toggleExportMenu() {
		showExportMenu = !showExportMenu;
	}


	function saveAsPDF() {
		showExportMenu = false;
		const html = editor?.getHTML() || "";
		const printWindow = window.open("", "_blank");
		if (printWindow) {
			printWindow.document.write(`
				<html>
				<head>
					<title>${title || 'Document'}</title>
					<style>
						body { font-family: 'Inter', sans-serif; padding: 40px; color: #1a0a12; line-height: 1.75; }
						h1 { font-size: 2em; margin-bottom: 0.5em; }
						h2 { font-size: 1.5em; margin-bottom: 0.5em; }
						p { margin-bottom: 1em; }
						ul, ol { margin-bottom: 1em; padding-left: 20px; }
						blockquote { border-left: 3px solid #a855f7; padding-left: 15px; color: #555; }
						pre { background: #f4f4f4; padding: 15px; border-radius: 5px; }
						.document-title { font-size: 2.5em; font-weight: bold; margin-bottom: 1em; text-align: center; }
					</style>
				</head>
				<body>
					<div class="document-title">${title || 'Untitled Document'}</div>
					${html}
				</body>
				</html>
			`);
			printWindow.document.close();
			printWindow.focus();
			setTimeout(() => {
				printWindow.print();
				printWindow.close();
			}, 250);
		}
	}

	function saveAsWord() {
		showExportMenu = false;
		const html = editor?.getHTML() || "";
		const header = "<html xmlns:o='urn:schemas-microsoft-com:office:office' xmlns:w='urn:schemas-microsoft-com:office:word' xmlns='http://www.w3.org/TR/REC-html40'><head><meta charset='utf-8'><title>" + (title || 'Document') + "</title></head><body>";
		const footer = "</body></html>";
		const sourceHTML = header + `<h1 style="text-align: center;">${title || 'Untitled Document'}</h1>` + html + footer;

		const blob = new Blob(['\ufeff', sourceHTML], {
			type: 'application/msword'
		});

		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = (title || 'document') + '.doc';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		URL.revokeObjectURL(url);
	}

	// Close export menu if clicked outside
	function handleOutsideClick(e: MouseEvent) {
		if (showExportMenu && !(e.target as Element).closest('.export-container')) {
			showExportMenu = false;
		}
	}

	async function runAI(action: string) {

		if (!editor || aiLoading) return;

		// Use the saved selection (in case editor lost focus)
		const sel = savedSelection ?? { from: editor.state.selection.from, to: editor.state.selection.to };
		const { from, to } = sel;
		if (from === to) return;

		const selectedText = editor.state.doc.textBetween(from, to, " ");
		if (!selectedText.trim()) return;

		aiLoading = action;
		try {
			const {
				data: { session },
			} = await supabase.auth.getSession();

			const res = await fetch(`${import.meta.env.VITE_API_URL}/ai/process`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					Authorization: `Bearer ${session?.access_token}`,
				},
				body: JSON.stringify({ text: selectedText, action }),
			});

			if (!res.ok) {
				console.error("AI request failed:", await res.text());
				return;
			}

			const { result } = await res.json();
			if (result) {
				// Convert markdown to HTML so TipTap renders tables, bold, etc.
				const html = marked.parse(result) as string;
				// Restore the saved selection before replacing
				editor
					.chain()
					.focus()
					.setTextSelection({ from, to })
					.deleteSelection()
					.insertContent(html)
					.run();
			}
		} catch (err) {
			console.error("AI error:", err);
		} finally {
			aiLoading = null;
			showBubble = false;
			savedSelection = null;
		}
	}

	async function askAI() {
		const prompt = aiCustomPrompt.trim();
		if (!editor || !prompt || aiLoading) return;

		// Use the saved selection (editor focus was lost when modal opened)
		const sel = savedSelection ?? { from: editor.state.selection.from, to: editor.state.selection.to };
		const { from, to } = sel;
		if (from === to) return;

		const selectedText = editor.state.doc.textBetween(from, to, " ");
		if (!selectedText.trim()) return;

		aiLoading = "custom";
		try {
			const {
				data: { session },
			} = await supabase.auth.getSession();

			const res = await fetch(`${import.meta.env.VITE_API_URL}/ai/process`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					Authorization: `Bearer ${session?.access_token}`,
				},
				body: JSON.stringify({ text: selectedText, action: "custom", customPrompt: prompt }),
			});

			if (!res.ok) {
				console.error("AI request failed:", await res.text());
				return;
			}

			const { result } = await res.json();
			if (result) {
				const html = marked.parse(result) as string;
				// Restore the saved selection before replacing
				editor
					.chain()
					.focus()
					.setTextSelection({ from, to })
					.deleteSelection()
					.insertContent(html)
					.run();
			}
			aiCustomPrompt = "";
		} catch (err) {
			console.error("AI error:", err);
		} finally {
			aiLoading = null;
			showBubble = false;
			showAskModal = false;
			savedSelection = null;
		}
	}

	function openAskModal() {
		// savedSelection is already set from when the bubble appeared.
		// Close the bubble and open the modal — no focus-loss risk.
		showBubble = false;
		showAskModal = true;
	}

	// ── Image upload helpers ──────────────────────────────────────
	function triggerImageUpload() { imageFileInput?.click(); }

	async function uploadCompressedImage(file: File): Promise<string> {
		const compressed = await imageCompression(file, {
			maxSizeMB: 1, maxWidthOrHeight: 1200, useWebWorker: true, fileType: file.type,
		});
		const ext = compressed.name.split('.').pop() ?? 'jpg';
		const filename = `doc-images/${docId}/${Date.now()}.${ext}`;
		const { error: uploadErr } = await supabase.storage.from('cascade-images')
			.upload(filename, compressed, { cacheControl: '3600', upsert: false, contentType: compressed.type });
		if (uploadErr) throw uploadErr;
		const { data: urlData } = supabase.storage.from('cascade-images').getPublicUrl(filename);
		if (!urlData.publicUrl) throw new Error('Failed to get public URL');
		return urlData.publicUrl;
	}

	async function handleImageFile(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		imageUploading = true; imageUploadError = null; imageUploadToast = 'compressing';
		try {
			imageUploadToast = 'uploading';
			const publicUrl = await uploadCompressedImage(file);
			editor?.chain().focus().setImage({ src: publicUrl, alt: file.name }).run();
			imageUploadToast = 'done';
			setTimeout(() => { imageUploadToast = 'idle'; }, 2500);
		} catch (err: unknown) {
			console.error('Image upload error:', err);
			imageUploadError = err instanceof Error ? err.message : 'Upload failed';
			imageUploadToast = 'error';
			setTimeout(() => { imageUploadToast = 'idle'; imageUploadError = null; }, 4000);
		} finally { imageUploading = false; }
	}

	async function replaceImage(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file || !editor) return;
		if (getImgPos() === null) return;
		imageUploading = true; imageUploadToast = 'compressing';
		try {
			imageUploadToast = 'uploading';
			const publicUrl = await uploadCompressedImage(file);
			editor.chain().updateAttributes('image', { src: publicUrl, alt: file.name }).run();
			imageUploadToast = 'done';
			setTimeout(() => { imageUploadToast = 'idle'; }, 2500);
		} catch (err: unknown) {
			console.error('Replace image error:', err);
			imageUploadError = err instanceof Error ? err.message : 'Upload failed';
			imageUploadToast = 'error';
			setTimeout(() => { imageUploadToast = 'idle'; imageUploadError = null; }, 4000);
		} finally { imageUploading = false; }
	}
</script>


<svelte:window on:click={handleOutsideClick} />

<svelte:head>
	<title>{title || "Untitled"} — Cascade</title>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link
		rel="preconnect"
		href="https://fonts.gstatic.com"
		crossorigin="anonymous"
	/>
	<link
		href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div class="layout">
	<aside class="sidebar">
		<div class="sidebar-inner">
			<a href="/dashboard" class="back-link">← Dashboard</a>

			{#if isImageSelected}
			<!-- ━━ IMAGE EDITING PANEL ━━━━━━━━━━━━━━━━━━━━━━ -->
			<div class="img-panel-header">
				<span class="img-panel-title">🖼 Image</span>
				<button class="img-panel-delete" onclick={deleteImage} title="Delete image">🗑</button>
			</div>

			<!-- Replace -->
			<div class="toolbar-section">
				<input type="file" accept="image/*" class="img-file-input" bind:this={imgReplaceInput} onchange={replaceImage} />
				<button class="tool-btn img-insert-btn" disabled={imageUploading} onclick={() => imgReplaceInput?.click()}>
					{#if imageUploading}<span class="ai-spin">⟳</span> Uploading…{:else}🔄 Replace Image{/if}
				</button>
			</div>

			<!-- Position -->
			<div class="toolbar-section">
				<span class="section-label">Position</span>
				<div class="btn-group">
					<button class="tool-btn" onclick={() => setImagePosition('left')}   title="Float left">◧ L</button>
					<button class="tool-btn" onclick={() => setImagePosition('center')} title="Center">⬛ C</button>
					<button class="tool-btn" onclick={() => setImagePosition('right')}  title="Float right">◨ R</button>
					<button class="tool-btn" onclick={() => setImagePosition('block')}  title="Block">▢ B</button>
				</div>
				<p class="img-hint">Drag image to reposition</p>
			</div>

			<!-- Width -->
			<div class="toolbar-section">
				<span class="section-label">Width — {imgWidthPct}%</span>
				<input type="range" class="img-slider" min="10" max="100" step="5"
					bind:value={imgWidthPct}
					oninput={applyImageStyle}
				/>
				<div class="btn-group">
					<button class="tool-btn" onclick={() => { imgWidthPct=25;  applyImageStyle(); }}>25%</button>
					<button class="tool-btn" onclick={() => { imgWidthPct=50;  applyImageStyle(); }}>50%</button>
					<button class="tool-btn" onclick={() => { imgWidthPct=75;  applyImageStyle(); }}>75%</button>
					<button class="tool-btn" onclick={() => { imgWidthPct=100; applyImageStyle(); }}>100%</button>
				</div>
			</div>

			<!-- Rotate -->
			<div class="toolbar-section">
				<span class="section-label">Rotate — {imgRotate}°</span>
				<input type="range" class="img-slider" min="-180" max="180" step="1"
					bind:value={imgRotate}
					oninput={applyImageStyle}
				/>
				<button class="tool-btn" style="width:100%" onclick={() => { imgRotate=0; applyImageStyle(); }}>Reset</button>
			</div>

			<!-- Border Radius -->
			<div class="toolbar-section">
				<span class="section-label">Corners — {imgRadius}px</span>
				<input type="range" class="img-slider" min="0" max="50" step="1"
					bind:value={imgRadius}
					oninput={applyImageStyle}
				/>
			</div>

			<!-- Filters -->
			<div class="toolbar-section">
				<span class="section-label">Brightness — {imgBrightness}%</span>
				<input type="range" class="img-slider" min="10" max="200" step="5"
					bind:value={imgBrightness} oninput={applyImageStyle}/>
			</div>
			<div class="toolbar-section">
				<span class="section-label">Contrast — {imgContrast}%</span>
				<input type="range" class="img-slider" min="10" max="200" step="5"
					bind:value={imgContrast} oninput={applyImageStyle}/>
			</div>
			<div class="toolbar-section">
				<span class="section-label">Saturation — {imgSaturation}%</span>
				<input type="range" class="img-slider" min="0" max="200" step="5"
					bind:value={imgSaturation} oninput={applyImageStyle}/>
			</div>
			<div class="toolbar-section">
				<span class="section-label">Blur — {imgBlur}px</span>
				<input type="range" class="img-slider" min="0" max="20" step="0.5"
					bind:value={imgBlur} oninput={applyImageStyle}/>
			</div>
			<div class="toolbar-section">
				<span class="section-label">Opacity — {imgOpacity}%</span>
				<input type="range" class="img-slider" min="10" max="100" step="5"
					bind:value={imgOpacity} oninput={applyImageStyle}/>
			</div>
			<div class="toolbar-section">
				<span class="section-label">Grayscale — {imgGrayscale}%</span>
				<input type="range" class="img-slider" min="0" max="100" step="5"
					bind:value={imgGrayscale} oninput={applyImageStyle}/>
			</div>
			<div class="toolbar-section">
				<span class="section-label">Sepia — {imgSepia}%</span>
				<input type="range" class="img-slider" min="0" max="100" step="5"
					bind:value={imgSepia} oninput={applyImageStyle}/>
			</div>

			<!-- Quick filter presets -->
			<div class="toolbar-section">
				<span class="section-label">Presets</span>
				<div class="btn-group" style="flex-wrap:wrap">
					<button class="tool-btn" onclick={() => { imgBrightness=100;imgContrast=100;imgSaturation=100;imgBlur=0;imgGrayscale=0;imgSepia=0;imgOpacity=100; applyImageStyle(); }}>Normal</button>
					<button class="tool-btn" onclick={() => { imgGrayscale=100;imgSepia=0; applyImageStyle(); }}>B&W</button>
					<button class="tool-btn" onclick={() => { imgSepia=80;imgGrayscale=0; applyImageStyle(); }}>Sepia</button>
					<button class="tool-btn" onclick={() => { imgBrightness=110;imgContrast=120;imgSaturation=130; applyImageStyle(); }}>Vivid</button>
					<button class="tool-btn" onclick={() => { imgBrightness=90;imgContrast=90;imgSaturation=80; applyImageStyle(); }}>Muted</button>
					<button class="tool-btn" onclick={() => { imgBlur=4; applyImageStyle(); }}>Soft</button>
				</div>
			</div>

			<!-- Alt text -->
			<div class="toolbar-section">
				<span class="section-label">Alt Text</span>
				<div style="display:flex;gap:4px">
					<input class="select-input" style="flex:1" bind:value={selectedImgAlt} placeholder="Describe image…" />
					<button class="tool-btn" onclick={() => { if(editor) editor.chain().focus().updateAttributes('image',{alt:selectedImgAlt}).run(); }}>✓</button>
				</div>
			</div>

			{:else}
			<!-- ━━ TEXT FORMATTING TOOLS ━━━━━━━━━━━━━━━━━━━━ -->

			<div class="toolbar-section">
				<span class="section-label">Format</span>
				<div class="btn-group">
					<button
						class="tool-btn"
						class:active={isBold}
						onclick={() =>
							editor?.chain().focus().toggleBold().run()}
						title="Bold"
					>
						<strong>B</strong>
					</button>
					<button
						class="tool-btn"
						class:active={isItalic}
						onclick={() =>
							editor?.chain().focus().toggleItalic().run()}
						title="Italic"
					>
						<em>I</em>
					</button>
					<button
						class="tool-btn"
						class:active={isUnderline}
						onclick={() =>
							editor?.chain().focus().toggleUnderline().run()}
						title="Underline"
					>
						<span style="text-decoration:underline">U</span>
					</button>
					<button
						class="tool-btn"
						class:active={isStrike}
						onclick={() =>
							editor?.chain().focus().toggleStrike().run()}
						title="Strikethrough"
					>
						<span style="text-decoration:line-through">S</span>
					</button>
				</div>
			</div>

			<div class="toolbar-section">
				<span class="section-label">Headings</span>
				<div class="btn-group">
					<button
						class="tool-btn heading-btn"
						class:active={isH1}
						onclick={() =>
							editor
								?.chain()
								.focus()
								.toggleHeading({ level: 1 })
								.run()}
						title="Heading 1">H1</button
					>
					<button
						class="tool-btn heading-btn"
						class:active={isH2}
						onclick={() =>
							editor
								?.chain()
								.focus()
								.toggleHeading({ level: 2 })
								.run()}
						title="Heading 2">H2</button
					>
					<button
						class="tool-btn heading-btn"
						class:active={isH3}
						onclick={() =>
							editor
								?.chain()
								.focus()
								.toggleHeading({ level: 3 })
								.run()}
						title="Heading 3">H3</button
					>
				</div>
			</div>

			<div class="toolbar-section">
				<span class="section-label">Font Size</span>
				<select
					class="select-input"
					value={currentFontSize}
					onchange={(e) =>
						applyFontSize((e.target as HTMLSelectElement).value)}
				>
					{#each fontSizes as size}
						<option value={size}>{size}</option>
					{/each}
				</select>
			</div>

			<div class="toolbar-section">
				<span class="section-label">Text Color</span>
				<div class="color-row">
					<input
						type="color"
						class="color-input"
						value={currentColor}
						oninput={(e) =>
							applyColor((e.target as HTMLInputElement).value)}
						title="Text color"
					/>
					<span class="color-label">{currentColor}</span>
				</div>
			</div>

			<div class="toolbar-section">
				<span class="section-label">Alignment</span>
				<div class="btn-group">
					<button
						class="tool-btn align-btn"
						class:active={isAlignLeft}
						onclick={() =>
							editor?.chain().focus().setTextAlign("left").run()}
						title="Align left">≡</button
					>
					<button
						class="tool-btn align-btn"
						class:active={isAlignCenter}
						onclick={() =>
							editor
								?.chain()
								.focus()
								.setTextAlign("center")
								.run()}
						title="Align center">≡</button
					>
					<button
						class="tool-btn align-btn"
						class:active={isAlignRight}
						onclick={() =>
							editor?.chain().focus().setTextAlign("right").run()}
						title="Align right">≡</button
					>
					<button
						class="tool-btn align-btn"
						class:active={isAlignJustify}
						onclick={() =>
							editor
								?.chain()
								.focus()
								.setTextAlign("justify")
								.run()}
						title="Justify">≡</button
					>
				</div>
			</div>

			<div class="toolbar-section">
				<span class="section-label">Lists</span>
				<div class="btn-group">
					<button
						class="tool-btn list-btn"
						class:active={isBulletList}
						onclick={() =>
							editor?.chain().focus().toggleBulletList().run()}
						title="Bullet list">• —</button
					>
					<button
						class="tool-btn list-btn"
						class:active={isOrderedList}
						onclick={() =>
							editor?.chain().focus().toggleOrderedList().run()}
						title="Ordered list">1. —</button
					>
				</div>
			</div>

			<div class="toolbar-section">
				<span class="section-label">History</span>
				<div class="btn-group">
					<button class="tool-btn" onclick={() => editor?.chain().focus().undo().run()} title="Undo">↩</button>
					<button class="tool-btn" onclick={() => editor?.chain().focus().redo().run()} title="Redo">↪</button>
				</div>
			</div>
			{/if}
		</div>
	</aside>

	<main class="panel">
		<header class="panel-header">
			<input
				class="title-input"
				type="text"
				bind:value={title}
				oninput={scheduleSave}
				placeholder="Untitled Document"
			/>
			
			<div class="header-actions">
				<span
					class="save-status"
					class:saving={saveStatus === "Saving..."}
					class:unsaved={saveStatus === "Unsaved"}
					class:offline={saveStatus === "Offline"}
				>
					{#if saveStatus === "Offline"}
						<span class="offline-dot"></span> Offline
					{:else if saveStatus === "Saving..."}
						<span class="ai-spin">⟳</span> Saving…
					{:else}
						{saveStatus}
					{/if}
				</span>

				<!-- Hidden file input lives here in the header -->
				<input
					type="file"
					accept="image/*"
					class="img-file-input"
					bind:this={imageFileInput}
					onchange={handleImageFile}
				/>
				<button
					class="header-img-btn"
					class:header-img-btn-loading={imageUploading}
					disabled={imageUploading}
					onclick={triggerImageUpload}
					title="Insert image"
				>
					{#if imageUploading}
						<span class="ai-spin">⟳</span> Uploading…
					{:else}
						🖼 Insert Image
					{/if}
				</button>

				<div class="export-container">
					<button class="export-btn" onclick={toggleExportMenu}>
						Export ▾
					</button>
					{#if showExportMenu}
						<div class="export-dropdown">
							<button class="export-option" onclick={saveAsPDF}>
								Save as PDF
							</button>
							<button class="export-option" onclick={saveAsWord}>
								Save as Word
							</button>
						</div>
					{/if}
				</div>
			</div>
		</header>

		<div class="editor-wrapper" data-cursor-text
			onmousedown={onEditorMouseDown}
		>
			<div class="editor-el" bind:this={editorEl}></div>
		</div>
	</main>
</div>

<!-- ── Image upload toast ── -->
{#if imageUploadToast !== 'idle'}
	<div
		class="img-upload-toast"
		class:toast-compressing={imageUploadToast === 'compressing'}
		class:toast-uploading={imageUploadToast === 'uploading'}
		class:toast-done={imageUploadToast === 'done'}
		class:toast-error={imageUploadToast === 'error'}
		role="status"
		aria-live="polite"
	>
		{#if imageUploadToast === 'compressing'}
			<span class="ai-spin">⟳</span> Compressing image…
		{:else if imageUploadToast === 'uploading'}
			<span class="ai-spin">⟳</span> Uploading to cloud…
		{:else if imageUploadToast === 'done'}
			✓ Image inserted!
		{:else if imageUploadToast === 'error'}
			⚠ {imageUploadError ?? 'Upload failed'}
		{/if}
	</div>
{/if}

<!-- ── Offline banner ── -->
{#if !isOnline}
	<div class="offline-banner" role="alert" aria-live="assertive">
		<span class="offline-banner-icon">📡</span>
		<div class="offline-banner-text">
			<strong>You're offline</strong>
			<span>Changes are saved locally and will sync when you reconnect.</span>
		</div>
	</div>
{/if}

<!-- ── Back-online toast ── -->
{#if isOnline && hasPendingOfflineSave}
	<div class="syncing-toast" role="status" aria-live="polite">
		<span class="ai-spin">⟳</span> Syncing offline changes…
	</div>
{/if}

<!-- ── Floating bubble menu ── -->
{#if showBubble && editor && !isImageSelected}
	<div
		class="bubble-menu"
		role="toolbar"
		aria-label="Text formatting"
		tabindex="-1"
		style="left: {bubbleX}px; top: {bubbleY}px;"
		onmousedown={(e) => e.preventDefault()}
	>
		<button
			class="bubble-arrow"
			onclick={() =>
				bubbleInner?.scrollBy({ left: -88, behavior: "smooth" })}
			title="Scroll left">‹</button
		>

		<div class="bubble-inner" bind:this={bubbleInner}>
			<button
				class="bb"
				class:on={isBold}
				onclick={() => editor?.chain().focus().toggleBold().run()}
				title="Bold"><strong>B</strong></button
			>

			<button
				class="bb"
				class:on={isItalic}
				onclick={() => editor?.chain().focus().toggleItalic().run()}
				title="Italic"><em>I</em></button
			>

			<button
				class="bb"
				class:on={isUnderline}
				onclick={() => editor?.chain().focus().toggleUnderline().run()}
				title="Underline"
				><span style="text-decoration:underline">U</span></button
			>

			<button
				class="bb"
				class:on={isStrike}
				onclick={() => editor?.chain().focus().toggleStrike().run()}
				title="Strikethrough"
				><span style="text-decoration:line-through">S</span></button
			>

			<div class="bb-divider"></div>

			<button
				class="bb"
				class:on={isH1}
				onclick={() =>
					editor?.chain().focus().toggleHeading({ level: 1 }).run()}
				title="Heading 1">H1</button
			>

			<button
				class="bb"
				class:on={isH2}
				onclick={() =>
					editor?.chain().focus().toggleHeading({ level: 2 }).run()}
				title="Heading 2">H2</button
			>

			<button
				class="bb"
				class:on={isH3}
				onclick={() =>
					editor?.chain().focus().toggleHeading({ level: 3 }).run()}
				title="Heading 3">H3</button
			>

			<div class="bb-divider"></div>

			<button
				class="bb"
				class:on={isAlignLeft}
				onclick={() =>
					editor?.chain().focus().setTextAlign("left").run()}
				title="Align left">≡L</button
			>

			<button
				class="bb"
				class:on={isAlignCenter}
				onclick={() =>
					editor?.chain().focus().setTextAlign("center").run()}
				title="Align center">≡C</button
			>

			<button
				class="bb"
				class:on={isAlignRight}
				onclick={() =>
					editor?.chain().focus().setTextAlign("right").run()}
				title="Align right">≡R</button
			>

			<div class="bb-divider"></div>

			<button
				class="bb"
				class:on={isBulletList}
				onclick={() => editor?.chain().focus().toggleBulletList().run()}
				title="Bullet list">• —</button
			>

			<button
				class="bb"
				class:on={isOrderedList}
				onclick={() =>
					editor?.chain().focus().toggleOrderedList().run()}
				title="Ordered list">1. —</button
			>
		</div>

		<button
			class="bubble-arrow"
			onclick={() =>
				bubbleInner?.scrollBy({ left: 88, behavior: "smooth" })}
			title="Scroll right">›</button
		>
	</div>
{/if}

<!-- ── AI Bubble (separate tooltip above the format bubble) ── -->
{#if showBubble && editor && !isImageSelected}
	<div
		class="ai-bubble"
		role="toolbar"
		aria-label="AI actions"
		tabindex="-1"
		style="left: {bubbleX}px; top: {bubbleY - 54}px;"
		onmousedown={(e) => e.preventDefault()}
	>
		<button
			class="ai-bubble-btn"
			class:ai-bbl-loading={aiLoading === 'fix grammar'}
			disabled={!!aiLoading}
			onclick={() => runAI('fix grammar')}
			title="Fix grammar"
		>
			{#if aiLoading === 'fix grammar'}
				<span class="ai-spin">⟳</span>
			{:else}
				✦ Fix
			{/if}
		</button>

		<button
			class="ai-bubble-btn"
			class:ai-bbl-loading={aiLoading === 'translate to Hindi'}
			disabled={!!aiLoading}
			onclick={() => runAI('translate to Hindi')}
			title="Translate to Hindi"
		>
			{#if aiLoading === 'translate to Hindi'}
				<span class="ai-spin">⟳</span>
			{:else}
				✦ HI
			{/if}
		</button>

		<button
			class="ai-bubble-btn"
			class:ai-bbl-loading={aiLoading === 'make a table'}
			disabled={!!aiLoading}
			onclick={() => runAI('make a table')}
			title="Make a table"
		>
			{#if aiLoading === 'make a table'}
				<span class="ai-spin">⟳</span>
			{:else}
				✦ ⊞
			{/if}
		</button>

		<button
			class="ai-bubble-btn"
			class:ai-bbl-loading={aiLoading === 'summarise'}
			disabled={!!aiLoading}
			onclick={() => runAI('summarise')}
			title="Summarise"
		>
			{#if aiLoading === 'summarise'}
				<span class="ai-spin">⟳</span>
			{:else}
				✦ ∑
			{/if}
		</button>

		<div class="ai-bubble-sep"></div>

		<button
			class="ai-bubble-btn ai-ask-open-btn"
			disabled={!!aiLoading}
			onclick={openAskModal}
			title="Ask AI a custom question"
		>✦ Ask AI</button>
	</div>
{/if}


<!-- ── Ask AI Modal ── -->

{#if showAskModal}
	<div
		class="ask-modal-overlay"
		onclick={() => { showAskModal = false; aiCustomPrompt = ""; }}
		onkeydown={(e) => e.key === 'Escape' && (showAskModal = false)}
		role="dialog"
		aria-modal="true"
		aria-label="Ask AI"
		tabindex="-1"
	>
		<div
			class="ask-modal"
			onclick={(e) => e.stopPropagation()}
			role="presentation"
		>
			<div class="ask-modal-header">
				<span class="ask-modal-title">✦ Ask AI</span>
				<button
					class="ask-modal-close"
					onclick={() => { showAskModal = false; aiCustomPrompt = ""; }}
					aria-label="Close"
				>✕</button>
			</div>
			<p class="ask-modal-hint">Describe what you'd like to do with the selected text.</p>
			<textarea
				class="ask-modal-input"
				placeholder="e.g. This content is code, clean up the formatting…"
				bind:value={aiCustomPrompt}
				disabled={!!aiLoading}
				rows="3"
				onkeydown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); askAI(); } }}
			></textarea>
			<div class="ask-modal-actions">
				<button
					class="ask-modal-cancel"
					onclick={() => { showAskModal = false; aiCustomPrompt = ""; }}
					disabled={!!aiLoading}
				>Cancel</button>
				<button
					class="ask-modal-submit"
					onclick={askAI}
					disabled={!!aiLoading || !aiCustomPrompt.trim()}
				>
					{#if aiLoading === 'custom'}
						<span class="ai-spin">⟳</span> Thinking…
					{:else}
						Apply
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(*, *::before, *::after) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}

	:global(body) {
		font-family: "Inter", sans-serif;
		background: #ffffff;
		color: #1a0a12;
	}

	.layout {
		display: flex;
		min-height: 100vh;
		width: 100%;
	}

	/* ── Sidebar ─────────────────────────────────── */
	.sidebar {
		width: 16%;
		min-width: 180px;
		background: #f8f8f8;
		border-right: 1px solid #ede8f0;
		display: flex;
		flex-direction: column;
		position: sticky;
		top: 0;
		height: 100vh;
		overflow-y: auto;
	}

	.sidebar-inner {
		padding: 20px 14px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.back-link {
		display: inline-block;
		font-size: 12px;
		color: #a855f7;
		text-decoration: none;
		font-weight: 500;
		margin-bottom: 16px;
		transition: color 0.15s;
	}
	.back-link:hover {
		color: #f472b4;
	}

	/* ── Toolbar sections ─────────────────────────── */
	.toolbar-section {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 10px 0;
		border-bottom: 1px solid #ede8f0;
	}
	.toolbar-section:last-child {
		border-bottom: none;
	}

	.section-label {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #b0a0bb;
	}

	.btn-group {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	/* ── Tool buttons ─────────────────────────────── */
	.tool-btn {
		min-width: 34px;
		height: 34px;
		padding: 0 8px;
		border: 1px solid #e2d9ee;
		border-radius: 7px;
		background: #ffffff;
		color: #3d2a50;
		font-family: "Inter", sans-serif;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 0.15s,
			border-color 0.15s,
			color 0.15s,
			box-shadow 0.15s;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.tool-btn:hover {
		background: #f3e8ff;
		border-color: #a855f7;
		color: #a855f7;
	}
	.tool-btn.active {
		background: linear-gradient(135deg, #f472b4 0%, #a855f7 100%);
		border-color: transparent;
		color: #ffffff;
		box-shadow: 0 2px 8px rgba(168, 85, 247, 0.35);
	}

	.heading-btn,
	.list-btn,
	.align-btn {
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.04em;
	}

	/* ── Selects & color inputs ──────────────────── */
	.select-input {
		width: 100%;
		height: 34px;
		border: 1px solid #e2d9ee;
		border-radius: 7px;
		background: #ffffff;
		color: #3d2a50;
		font-family: "Inter", sans-serif;
		font-size: 13px;
		padding: 0 8px;
		cursor: pointer;
		outline: none;
		transition: border-color 0.15s;
	}
	.select-input:focus {
		border-color: #a855f7;
	}

	.color-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.color-input {
		width: 36px;
		height: 36px;
		border: 1px solid #e2d9ee;
		border-radius: 7px;
		padding: 2px;
		background: #fff;
		cursor: pointer;
	}
	.color-label {
		font-size: 11px;
		color: #8c7aa0;
		font-family: monospace;
	}

	/* ── Right panel ─────────────────────────────── */
	.panel {
		width: 80%;
		display: flex;
		flex-direction: column;
		background: #ffffff;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 18px 40px 14px;
		border-bottom: 1px solid #f0eaf7;
		position: sticky;
		top: 0;
		background: rgba(255, 255, 255, 0.92);
		backdrop-filter: blur(8px);
		z-index: 10;
	}

	.title-input {
		flex: 1;
		font-family: "Inter", sans-serif;
		font-size: 22px;
		font-weight: 700;
		color: #1a0a12;
		border: none;
		outline: none;
		background: transparent;
		margin-right: 20px;
	}
	.title-input::placeholder {
		color: #cdc0d9;
	}

	.save-status {
		font-size: 12px;
		font-weight: 500;
		color: #a0c890;
		white-space: nowrap;
		transition: color 0.2s;
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.save-status.saving {
		color: #a855f7;
	}
	.save-status.unsaved {
		color: #f472b4;
	}
	.save-status.offline {
		color: #fb923c;
		font-weight: 600;
	}
	/* Pulsing dot indicator for offline state */
	.offline-dot {
		display: inline-block;
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #fb923c;
		animation: offlinePulse 1.4s ease-in-out infinite;
		flex-shrink: 0;
	}
	@keyframes offlinePulse {
		0%, 100% { opacity: 1;   transform: scale(1); }
		50%       { opacity: 0.4; transform: scale(0.75); }
	}

	/* ── Export Dropdown ─────────────────────────── */
	.header-actions {
		display: flex;
		align-items: center;
		gap: 20px;
	}

	.export-container {
		position: relative;
	}

	.export-btn {
		padding: 8px 16px;
		font-size: 13px;
		font-weight: 600;
		border: 1px solid #e2d9ee;
		border-radius: 8px;
		background: #fff;
		color: #3d2a50;
		cursor: pointer;
		transition: all 0.15s;
	}
	.export-btn:hover {
		background: #f3e8ff;
		border-color: #a855f7;
		color: #a855f7;
	}

	.export-dropdown {
		position: absolute;
		top: calc(100% + 8px);
		right: 0;
		width: 140px;
		background: #ffffff;
		border: 1px solid #e2d9ee;
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
		display: flex;
		flex-direction: column;
		padding: 6px;
		z-index: 50;
	}

	.export-option {
		padding: 8px 12px;
		font-size: 13px;
		font-family: "Inter", sans-serif;
		font-weight: 500;
		color: #1a0a12;
		background: transparent;
		border: none;
		border-radius: 6px;
		text-align: left;
		cursor: pointer;
		transition: background 0.1s, color 0.1s;
	}
	.export-option:hover {
		background: #f3e8ff;
		color: #a855f7;
	}

	/* ── Editor area ─────────────────────────────── */
	.editor-wrapper {
		flex: 1;
		padding: 32px 40px 60px;
	}

	.editor-el {
		min-height: 100vh;
		outline: none;
	}

	/* TipTap ProseMirror core styles */
	:global(.ProseMirror) {
		min-height: 100vh;
		outline: none;
		font-family: "Inter", sans-serif;
		font-size: 16px;
		line-height: 1.75;
		color: #1a0a12;
		caret-color: #a855f7;
	}

	:global(.ProseMirror h1) {
		font-size: 2em;
		font-weight: 700;
		margin: 0.6em 0 0.3em;
		color: #1a0a12;
	}
	:global(.ProseMirror h2) {
		font-size: 1.5em;
		font-weight: 600;
		margin: 0.6em 0 0.3em;
	}
	:global(.ProseMirror h3) {
		font-size: 1.2em;
		font-weight: 600;
		margin: 0.5em 0 0.25em;
	}
	:global(.ProseMirror p) {
		margin: 0.4em 0;
	}
	:global(.ProseMirror ul),
	:global(.ProseMirror ol) {
		padding-left: 1.6em;
		margin: 0.4em 0;
	}
	:global(.ProseMirror li) {
		margin: 0.2em 0;
	}
	:global(.ProseMirror strong) {
		font-weight: 700;
	}
	:global(.ProseMirror em) {
		font-style: italic;
	}
	:global(.ProseMirror u) {
		text-decoration: underline;
	}
	:global(.ProseMirror s) {
		text-decoration: line-through;
	}
	:global(.ProseMirror blockquote) {
		border-left: 3px solid #a855f7;
		margin: 0.8em 0;
		padding-left: 1em;
		color: #6b5b80;
	}
	:global(.ProseMirror code) {
		background: #f3e8ff;
		border-radius: 4px;
		padding: 0.15em 0.4em;
		font-size: 0.9em;
		color: #a855f7;
	}
	:global(.ProseMirror pre) {
		background: #1a0a12;
		color: #f3e8ff;
		border-radius: 8px;
		padding: 1em 1.2em;
		margin: 0.8em 0;
		overflow-x: auto;
	}
	:global(.ProseMirror pre code) {
		background: none;
		color: inherit;
		padding: 0;
	}
	/* Placeholder */
	:global(.ProseMirror p.is-editor-empty:first-child::before) {
		content: "Start writing…";
		color: #cdc0d9;
		pointer-events: none;
		float: left;
		height: 0;
	}

	/* ── Floating bubble menu ─────────────────────── */
	:global(.bubble-menu) {
		position: fixed;
		z-index: 9000;
		transform: translate(-50%, calc(-100% - 6px));
		display: flex;
		align-items: center;
		gap: 2px;
		background: rgba(18, 6, 28, 0.88);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		border: 1px solid rgba(168, 85, 247, 0.25);
		border-radius: 12px;
		padding: 4px 6px;
		box-shadow:
			0 6px 24px rgba(0, 0, 0, 0.35),
			0 0 0 1px rgba(244, 114, 182, 0.08);
		pointer-events: all;
		/* Small downward caret */
		filter: drop-shadow(0 2px 6px rgba(168, 85, 247, 0.2));
		animation: bubblePop 0.15s ease both;
	}

	:global(.bubble-menu::after) {
		content: "";
		position: absolute;
		bottom: -6px;
		left: 50%;
		transform: translateX(-50%);
		border: 6px solid transparent;
		border-top-color: rgba(18, 6, 28, 0.88);
		border-bottom: none;
	}

	/* Scrollable track — hides the scrollbar */
	:global(.bubble-inner) {
		display: flex;
		align-items: center;
		gap: 2px;
		overflow-x: auto;
		max-width: 420px;
		scrollbar-width: none;
	}
	:global(.bubble-inner::-webkit-scrollbar) {
		display: none;
	}

	/* Individual tool button */
	:global(.bb) {
		flex-shrink: 0;
		min-width: 30px;
		height: 30px;
		padding: 0 7px;
		border: none;
		border-radius: 7px;
		background: transparent;
		color: rgba(255, 255, 255, 0.75);
		font-family: "Inter", sans-serif;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
		transition:
			background 0.12s,
			color 0.12s;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	:global(.bb:hover) {
		background: rgba(168, 85, 247, 0.2);
		color: #fff;
	}
	:global(.bb.on) {
		background: linear-gradient(135deg, #f472b4 0%, #a855f7 100%);
		color: #fff;
		box-shadow: 0 0 8px rgba(168, 85, 247, 0.5);
	}

	/* Vertical divider between groups */
	:global(.bb-divider) {
		width: 1px;
		height: 18px;
		background: rgba(255, 255, 255, 0.12);
		flex-shrink: 0;
		margin: 0 2px;
	}

	/* ‹ › scroll navigation arrows */
	:global(.bubble-arrow) {
		flex-shrink: 0;
		width: 22px;
		height: 30px;
		border: none;
		border-radius: 6px;
		background: transparent;
		color: rgba(255, 255, 255, 0.45);
		font-size: 16px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition:
			color 0.12s,
			background 0.12s;
		padding: 0;
	}
	:global(.bubble-arrow:hover) {
		color: #f472b4;
		background: rgba(244, 114, 182, 0.12);
	}

	@keyframes bubblePop {
		from {
			opacity: 0;
			transform: translate(-50%, calc(-100% - 2px)) scale(0.95);
		}
		to {
			opacity: 1;
			transform: translate(-50%, calc(-100% - 6px)) scale(1);
		}
	}

	/* ── AI action buttons ─────────────────────── */
	:global(.ai-btn) {
		min-width: 38px;
		padding: 0 9px;
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.18) 0%, rgba(244, 114, 182, 0.12) 100%);
		color: #e879f9;
		border: 1px solid rgba(168, 85, 247, 0.3);
		border-radius: 7px;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.02em;
		white-space: nowrap;
		transition:
			background 0.15s,
			color 0.15s,
			box-shadow 0.15s,
			opacity 0.15s;
	}
	:global(.ai-btn:hover:not(:disabled)) {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.35) 0%, rgba(244, 114, 182, 0.25) 100%);
		color: #fff;
		box-shadow: 0 0 10px rgba(168, 85, 247, 0.4);
		border-color: rgba(168, 85, 247, 0.6);
	}
	:global(.ai-btn:disabled) {
		opacity: 0.55;
		cursor: not-allowed;
	}
	:global(.ai-btn.ai-loading) {
		background: linear-gradient(135deg, #a855f7 0%, #f472b4 100%);
		color: #fff;
		opacity: 1;
	}

	/* Spin animation for loading indicator */
	:global(.ai-spin) {
		display: inline-block;
		animation: aiSpin 0.8s linear infinite;
		font-size: 15px;
		line-height: 1;
	}
	@keyframes aiSpin {
		from { transform: rotate(0deg); }
		to   { transform: rotate(360deg); }
	}

	/* ── Table styles (for AI-generated tables) ─── */
	:global(.ProseMirror table) {
		border-collapse: collapse;
		width: 100%;
		margin: 1em 0;
		font-size: 14px;
		overflow: hidden;
		border-radius: 8px;
		border: 1px solid #e2d9ee;
	}
	:global(.ProseMirror th) {
		background: linear-gradient(135deg, #f3e8ff 0%, #fce7f3 100%);
		color: #3d2a50;
		font-weight: 600;
		padding: 10px 14px;
		text-align: left;
		border: 1px solid #e2d9ee;
		font-size: 12px;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	:global(.ProseMirror td) {
		padding: 9px 14px;
		border: 1px solid #ede8f0;
		color: #1a0a12;
		vertical-align: top;
		line-height: 1.5;
	}
	:global(.ProseMirror tr:nth-child(even) td) {
		background: #faf8fc;
	}
	:global(.ProseMirror tr:hover td) {
		background: #f3e8ff55;
	}

	/* ── Standalone AI bubble ─────────────────────── */
	:global(.ai-bubble) {
		position: fixed;
		z-index: 9001;
		transform: translate(-50%, calc(-100% - 6px));
		display: flex;
		align-items: center;
		gap: 4px;
		background: rgba(14, 4, 22, 0.92);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		border: 1px solid rgba(168, 85, 247, 0.45);
		border-radius: 14px;
		padding: 5px 8px;
		box-shadow:
			0 6px 28px rgba(0, 0, 0, 0.4),
			0 0 0 1px rgba(244, 114, 182, 0.1),
			0 0 18px rgba(168, 85, 247, 0.15);
		pointer-events: all;
		animation: bubblePop 0.15s ease both;
		white-space: nowrap;
	}

	/* Small upward-pointing caret under AI bubble */
	:global(.ai-bubble::after) {
		content: "";
		position: absolute;
		bottom: -6px;
		left: 50%;
		transform: translateX(-50%);
		border: 6px solid transparent;
		border-top-color: rgba(14, 4, 22, 0.92);
		border-bottom: none;
	}

	/* Quick-action buttons inside the AI bubble */
	:global(.ai-bubble-btn) {
		flex-shrink: 0;
		height: 28px;
		padding: 0 10px;
		border: 1px solid rgba(168, 85, 247, 0.35);
		border-radius: 8px;
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.15) 0%, rgba(244, 114, 182, 0.08) 100%);
		color: #e879f9;
		font-family: "Inter", sans-serif;
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.04em;
		cursor: pointer;
		transition: background 0.13s, color 0.13s, box-shadow 0.13s, opacity 0.13s;
		display: flex;
		align-items: center;
		gap: 4px;
	}
	:global(.ai-bubble-btn:hover:not(:disabled)) {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.4) 0%, rgba(244, 114, 182, 0.25) 100%);
		color: #fff;
		border-color: rgba(168, 85, 247, 0.7);
		box-shadow: 0 0 12px rgba(168, 85, 247, 0.4);
	}
	:global(.ai-bubble-btn:disabled) {
		opacity: 0.5;
		cursor: not-allowed;
	}
	:global(.ai-bubble-btn.ai-bbl-loading) {
		background: linear-gradient(135deg, #a855f7 0%, #f472b4 100%);
		color: #fff;
		border-color: transparent;
		opacity: 1;
	}

	/* Vertical divider between quick-buttons and Ask AI */
	:global(.ai-bubble-sep) {
		width: 1px;
		height: 20px;
		background: rgba(168, 85, 247, 0.3);
		flex-shrink: 0;
		margin: 0 2px;
	}

	/* Ask AI open button (inside bubble) */
	:global(.ai-ask-open-btn) {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.25) 0%, rgba(244, 114, 182, 0.15) 100%) !important;
		border-color: rgba(168, 85, 247, 0.5) !important;
		color: #fff !important;
	}
	:global(.ai-ask-open-btn:hover:not(:disabled)) {
		background: linear-gradient(135deg, #a855f7 0%, #f472b4 100%) !important;
		box-shadow: 0 0 14px rgba(168, 85, 247, 0.5) !important;
	}

	/* ── Ask AI Modal ─────────────────────────────── */
	.ask-modal-overlay {
		position: fixed;
		inset: 0;
		z-index: 10000;
		background: rgba(10, 4, 18, 0.6);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		animation: fadeIn 0.15s ease;
	}
	@keyframes fadeIn {
		from { opacity: 0; }
		to   { opacity: 1; }
	}
	.ask-modal {
		background: #ffffff;
		border-radius: 16px;
		width: 480px;
		max-width: 92vw;
		padding: 28px;
		box-shadow:
			0 24px 60px rgba(0, 0, 0, 0.18),
			0 0 0 1px rgba(168, 85, 247, 0.12);
		animation: modalPop 0.2s cubic-bezier(0.34, 1.56, 0.64, 1) both;
	}
	@keyframes modalPop {
		from { opacity: 0; transform: scale(0.92) translateY(8px); }
		to   { opacity: 1; transform: scale(1) translateY(0); }
	}
	.ask-modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 8px;
	}
	.ask-modal-title {
		font-size: 16px;
		font-weight: 700;
		color: #3d2a50;
		background: linear-gradient(135deg, #a855f7, #f472b4);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}
	.ask-modal-close {
		width: 28px;
		height: 28px;
		border: none;
		background: #f3e8ff;
		color: #a855f7;
		border-radius: 8px;
		cursor: pointer;
		font-size: 13px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.15s;
	}
	.ask-modal-close:hover { background: #e9d5ff; }
	.ask-modal-hint {
		font-size: 12px;
		color: #8c7aa0;
		margin-bottom: 14px;
	}
	.ask-modal-input {
		width: 100%;
		padding: 12px 14px;
		border: 1.5px solid #e2d9ee;
		border-radius: 10px;
		background: #faf8fc;
		color: #1a0a12;
		font-family: "Inter", sans-serif;
		font-size: 14px;
		line-height: 1.5;
		resize: vertical;
		min-height: 80px;
		outline: none;
		transition: border-color 0.15s, box-shadow 0.15s;
		margin-bottom: 16px;
	}
	.ask-modal-input:focus {
		border-color: #a855f7;
		box-shadow: 0 0 0 3px rgba(168, 85, 247, 0.12);
	}
	.ask-modal-input:disabled { opacity: 0.6; }
	.ask-modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
	}
	.ask-modal-cancel {
		padding: 9px 18px;
		border: 1px solid #e2d9ee;
		border-radius: 9px;
		background: #fff;
		color: #8c7aa0;
		font-family: "Inter", sans-serif;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.15s;
	}
	.ask-modal-cancel:hover { background: #f3e8ff; color: #a855f7; }
	.ask-modal-cancel:disabled { opacity: 0.5; cursor: not-allowed; }
	.ask-modal-submit {
		padding: 9px 22px;
		border: none;
		border-radius: 9px;
		background: linear-gradient(135deg, #a855f7 0%, #f472b4 100%);
		color: #fff;
		font-family: "Inter", sans-serif;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: opacity 0.15s, box-shadow 0.15s;
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.ask-modal-submit:hover:not(:disabled) {
		box-shadow: 0 4px 14px rgba(168, 85, 247, 0.4);
	}
	.ask-modal-submit:disabled { opacity: 0.45; cursor: not-allowed; }

	/* ── Hidden file input ──────────────────────── */
	.img-file-input {
		display: none;
	}

	/* ── Header Insert Image button ─────────────────── */
	.header-img-btn {
		padding: 8px 16px;
		font-size: 13px;
		font-weight: 600;
		border: 1px solid rgba(168, 85, 247, 0.45);
		border-radius: 8px;
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.08) 0%, rgba(244, 114, 182, 0.05) 100%);
		color: #7c3aed;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 6px;
		white-space: nowrap;
		transition: all 0.15s;
	}
	.header-img-btn:hover:not(:disabled) {
		background: linear-gradient(135deg, rgba(168, 85, 247, 0.2) 0%, rgba(244, 114, 182, 0.12) 100%);
		border-color: #a855f7;
		color: #a855f7;
		box-shadow: 0 2px 10px rgba(168, 85, 247, 0.25);
	}
	.header-img-btn-loading,
	.header-img-btn:disabled {
		opacity: 0.65;
		cursor: not-allowed;
	}

	/* ── ProseMirror image ───────────────────── */
	:global(.ProseMirror img) {
		max-width: 100%;
		height: auto;
		border-radius: 10px;
		margin: 1em 0;
		display: block;
		border: 2px solid transparent;
		transition: border-color 0.15s, box-shadow 0.15s;
		cursor: default;
	}
	:global(.ProseMirror img.ProseMirror-selectednode) {
		border-color: #a855f7;
		box-shadow: 0 0 0 3px rgba(168, 85, 247, 0.2);
	}

	/* ── Upload toast ──────────────────────── */
	.img-upload-toast {
		position: fixed;
		bottom: 28px;
		right: 28px;
		z-index: 20000;
		padding: 12px 20px;
		border-radius: 12px;
		font-family: "Inter", sans-serif;
		font-size: 13px;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 8px;
		background: rgba(18, 6, 28, 0.9);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		border: 1px solid rgba(168, 85, 247, 0.3);
		color: #e2ceff;
		box-shadow: 0 8px 28px rgba(0, 0, 0, 0.35);
		animation: toastSlideIn 0.2s cubic-bezier(0.34, 1.56, 0.64, 1) both;
	}
	.toast-done {
		color: #86efac;
		border-color: rgba(134, 239, 172, 0.4);
	}
	.toast-error {
		color: #fca5a5;
		border-color: rgba(252, 165, 165, 0.4);
	}
	@keyframes toastSlideIn {
		from { opacity: 0; transform: translateY(12px) scale(0.95); }
		to   { opacity: 1; transform: translateY(0) scale(1); }
	}

	/* \u2500\u2500 Image editing sidebar panel \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500 */
	.img-panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 6px 0 10px;
		border-bottom: 1px solid #ede8f0;
		margin-bottom: 4px;
	}
	.img-panel-title {
		font-size: 13px;
		font-weight: 700;
		background: linear-gradient(135deg, #a855f7, #f472b4);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}
	.img-panel-delete {
		width: 28px;
		height: 28px;
		border: 1px solid rgba(252, 165, 165, 0.35);
		border-radius: 7px;
		background: rgba(252, 165, 165, 0.08);
		color: #fca5a5;
		font-size: 14px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.15s, border-color 0.15s;
	}
	.img-panel-delete:hover {
		background: rgba(252, 165, 165, 0.2);
		border-color: rgba(252, 165, 165, 0.6);
	}

	/* Slider track */
	.img-slider {
		width: 100%;
		height: 4px;
		-webkit-appearance: none;
		appearance: none;
		border-radius: 4px;
		background: linear-gradient(to right, #a855f7, #f472b4);
		outline: none;
		cursor: pointer;
		margin: 6px 0 4px;
	}
	.img-slider::-webkit-slider-thumb {
		-webkit-appearance: none;
		width: 14px;
		height: 14px;
		border-radius: 50%;
		background: #fff;
		border: 2px solid #a855f7;
		box-shadow: 0 1px 4px rgba(168, 85, 247, 0.4);
		cursor: grab;
	}
	.img-slider::-webkit-slider-thumb:active { cursor: grabbing; }
	.img-slider::-moz-range-thumb {
		width: 14px;
		height: 14px;
		border-radius: 50%;
		background: #fff;
		border: 2px solid #a855f7;
		cursor: grab;
	}

	.img-insert-btn {
		width: 100%;
		justify-content: center;
		gap: 6px;
		font-size: 12px;
	}

	/* Hint text under position buttons */
	.img-hint {
		font-size: 10px;
		color: #b0a0bb;
		margin-top: 4px;
		font-style: italic;
	}

	/* Draggable cursor on selected image */
	:global(.ProseMirror img.ProseMirror-selectednode) {
		cursor: grab !important;
		outline: 3px solid #a855f7;
		outline-offset: 3px;
		box-shadow: 0 0 0 6px rgba(168, 85, 247, 0.12);
		border-radius: 8px;
	}
	:global(.ProseMirror img.ProseMirror-selectednode:active) {
		cursor: grabbing !important;
	}

	/* ── Offline banner ────────────────────────────────────── */
	.offline-banner {
		position: fixed;
		bottom: 0;
		left: 0;
		right: 0;
		z-index: 30000;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 28px;
		background: linear-gradient(90deg, #431407 0%, #7c2d12 50%, #431407 100%);
		border-top: 1px solid rgba(251, 146, 60, 0.4);
		box-shadow: 0 -4px 24px rgba(251, 146, 60, 0.2);
		animation: bannerSlideUp 0.3s cubic-bezier(0.34, 1.56, 0.64, 1) both;
	}
	@keyframes bannerSlideUp {
		from { transform: translateY(100%); opacity: 0; }
		to   { transform: translateY(0);    opacity: 1; }
	}
	.offline-banner-icon {
		font-size: 20px;
		animation: offlinePulse 2s ease-in-out infinite;
		flex-shrink: 0;
	}
	.offline-banner-text {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.offline-banner-text strong {
		font-size: 13px;
		font-weight: 700;
		color: #fed7aa;
		font-family: "Inter", sans-serif;
		letter-spacing: 0.01em;
	}
	.offline-banner-text span {
		font-size: 11px;
		color: #fdba74;
		font-family: "Inter", sans-serif;
	}

	/* ── Syncing toast (back-online) ─────────────────────── */
	.syncing-toast {
		position: fixed;
		bottom: 28px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 30001;
		padding: 11px 22px;
		border-radius: 12px;
		font-family: "Inter", sans-serif;
		font-size: 13px;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 8px;
		background: rgba(18, 6, 28, 0.9);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		border: 1px solid rgba(168, 85, 247, 0.4);
		color: #e2ceff;
		box-shadow: 0 8px 28px rgba(0, 0, 0, 0.4), 0 0 16px rgba(168, 85, 247, 0.2);
		animation: toastSlideIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1) both;
	}
</style>
