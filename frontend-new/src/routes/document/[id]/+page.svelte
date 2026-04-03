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
	import { supabase } from "$lib/supabase";

	let { data } = $props();

	const docId = $page.params.id;

	let editor = $state<Editor | null>(null);
	let editorEl: HTMLDivElement;
	let title = $state<string>(data.document.title ?? "");
	let saveStatus = $state<"Saved" | "Saving..." | "Unsaved">("Saved");
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

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

	function updateBubble(e: Editor) {
		const { from, to } = e.state.selection;
		if (from === to) {
			showBubble = false;
			return;
		}
		const sel = window.getSelection();
		if (!sel || sel.rangeCount === 0 || sel.isCollapsed) {
			showBubble = false;
			return;
		}
		const rect = sel.getRangeAt(0).getBoundingClientRect();
		if (!rect.width) {
			showBubble = false;
			return;
		}
		// Clamp X so bubble never clips the viewport edges
		const halfW = 160; // approx half of bubble width
		bubbleX = Math.min(
			Math.max(rect.left + rect.width / 2, halfW),
			window.innerWidth - halfW,
		);
		bubbleY = rect.top - 10; // 10px gap above selection
		showBubble = true;
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
		debounceTimer = setTimeout(saveDocument, 1500);
	}

	async function saveDocument() {
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

	onMount(() => {
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
</script>

<svelte:window on:click={handleOutsideClick} />

<svelte:head>
	<title>{title || "Untitled"} — Collaborent</title>
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
					<button
						class="tool-btn"
						onclick={() => editor?.chain().focus().undo().run()}
						title="Undo">↩</button
					>
					<button
						class="tool-btn"
						onclick={() => editor?.chain().focus().redo().run()}
						title="Redo">↪</button
					>
				</div>
			</div>
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
				>
					{saveStatus}
				</span>

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

		<div class="editor-wrapper" data-cursor-text>
			<div class="editor-el" bind:this={editorEl}></div>
		</div>
	</main>
</div>

<!-- ── Floating bubble menu ── -->
{#if showBubble && editor}
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
	}
	.save-status.saving {
		color: #a855f7;
	}
	.save-status.unsaved {
		color: #f472b4;
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
		max-width: 260px;
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
</style>
