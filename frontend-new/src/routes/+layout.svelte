<script lang="ts">
	import { onMount } from "svelte";

	let { children } = $props();

	let cursorX = $state(-400);
	let cursorY = $state(-400);
	let isVisible = $state(false);
	let isOverButton = $state(false);
	let isOverEditor = $state(false);
	let isMoving = $state(false);
	let rotAngle = $state(0); // degrees, direction of travel

	interface JellyDot {
		id: number;
		x: number;
		y: number;
		size: number; // px
		color: string; // rgba
		life: number; // ms
	}

	let dots: JellyDot[] = $state([]);
	let dotId = 0;

	// 7 tentacle lanes: angleOffset from behind-direction, maxDist, dot count, color
	const LANES = [
		{ off: -52, dist: 36, n: 2, color: "139, 92, 246" }, // outer-left  violet
		{ off: -32, dist: 52, n: 3, color: "168, 85, 247" }, // mid-left    purple
		{ off: -14, dist: 64, n: 3, color: "192, 72, 219" }, // inner-left  pink-purple
		{ off: 0, dist: 72, n: 4, color: "236, 72, 153" }, // center      pink
		{ off: 14, dist: 64, n: 3, color: "192, 72, 219" }, // inner-right pink-purple
		{ off: 32, dist: 52, n: 3, color: "168, 85, 247" }, // mid-right   purple
		{ off: 52, dist: 36, n: 2, color: "139, 92, 246" }, // outer-right violet
	];

	let prevX = -400,
		prevY = -400;
	let moveRad = 0; // raw movement angle in radians
	let targetDeg = 0; // target rotation in degrees
	let smoothDeg = 0; // interpolated rotation
	let movingTimer: ReturnType<typeof setTimeout>;
	let rafId: number;
	let lastSpawn = 0;

	onMount(() => {
		// Smooth rotation loop
		const tick = () => {
			let delta = targetDeg - smoothDeg;
			while (delta > 180) delta -= 360;
			while (delta < -180) delta += 360;
			smoothDeg += delta * 0.1;
			rotAngle = smoothDeg;
			rafId = requestAnimationFrame(tick);
		};
		rafId = requestAnimationFrame(tick);

		const handleMouseMove = (e: MouseEvent) => {
			const dx = e.clientX - prevX;
			const dy = e.clientY - prevY;
			const speed = Math.sqrt(dx * dx + dy * dy);

			cursorX = e.clientX;
			cursorY = e.clientY;
			if (!isVisible) isVisible = true;

			if (speed > 2.5) {
				moveRad = Math.atan2(dy, dx);
				// +90° so the "flat bottom" of the bell faces the trail
				targetDeg = moveRad * (180 / Math.PI) + 90;
				isMoving = true;
				clearTimeout(movingTimer);
				movingTimer = setTimeout(() => {
					isMoving = false;
				}, 140);
			}

			prevX = e.clientX;
			prevY = e.clientY;

			// Spawn dot-tentacles (suppressed while over a button or the editor)
			const now = Date.now();
			if (
				!isOverButton &&
				!isOverEditor &&
				speed > 3 &&
				now - lastSpawn > 18
			) {
				lastSpawn = now;
				const trailRad = moveRad + Math.PI; // opposite = behind

				for (const lane of LANES) {
					const laneRad = trailRad + (lane.off * Math.PI) / 180;
					const perpRad = laneRad + Math.PI / 2;

					for (let d = 0; d < lane.n; d++) {
						const t = (d + 1) / lane.n; // 0→1 (1=tip)
						const dist = lane.dist * t;
						const jitter = (Math.random() - 0.5) * 7;
						const dotX =
							e.clientX +
							Math.cos(laneRad) * dist +
							Math.cos(perpRad) * jitter;
						const dotY =
							e.clientY +
							Math.sin(laneRad) * dist +
							Math.sin(perpRad) * jitter;
						const size = 6.5 * (1 - t * 0.6); // 6.5px near bell → ~2.6px at tip
						// Long lifetime so dots are still alive when sphere reforms
						const life = 1050 + (1 - t) * 300;
						const id = dotId++;

						dots = [
							...dots,
							{
								id,
								x: dotX,
								y: dotY,
								size,
								color: lane.color,
								life,
							},
						];

						setTimeout(() => {
							dots = dots.filter((dot) => dot.id !== id);
						}, life);
					}
				}
			}
		};

		const handleMouseLeave = () => {
			isVisible = false;
		};

		// Button / input / select hover detection via event delegation
		const handleMouseOver = (e: MouseEvent) => {
			const t = e.target as Element;
			if (
				t.closest(
					'button, input, select, a[role="button"], [data-cursor-glow], card-body',
				)
			) {
				isOverButton = true;
			}
			if (t.closest("[data-cursor-text]")) {
				isOverEditor = true;
			}
		};
		const handleMouseOut = (e: MouseEvent) => {
			const t = e.target as Element;
			if (
				t.closest(
					'button, input, select, a[role="button"], [data-cursor-glow]',
				)
			) {
				isOverButton = false;
			}
			if (t.closest("[data-cursor-text]")) {
				isOverEditor = false;
			}
		};

		window.addEventListener("mousemove", handleMouseMove);
		window.addEventListener("mouseover", handleMouseOver);
		window.addEventListener("mouseout", handleMouseOut);
		document.documentElement.addEventListener(
			"mouseleave",
			handleMouseLeave,
		);

		return () => {
			window.removeEventListener("mousemove", handleMouseMove);
			window.removeEventListener("mouseover", handleMouseOver);
			window.removeEventListener("mouseout", handleMouseOut);
			document.documentElement.removeEventListener(
				"mouseleave",
				handleMouseLeave,
			);
			cancelAnimationFrame(rafId);
			clearTimeout(movingTimer);
		};
	});
</script>

<svelte:head>
	<link rel="icon" href="/organized-folder.png?v=2" />
</svelte:head>

{#if isVisible}
	<!-- Dot tentacle particles — hidden while over a button -->
	{#if !isOverButton && !isOverEditor}
		{#each dots as dot (dot.id)}
			<div
				class="jelly-dot"
				style="
					left: {dot.x}px;
					top: {dot.y}px;
					width: {dot.size}px;
					height: {dot.size}px;
					--c: {dot.color};
					animation-duration: {dot.life}ms;
				"
			></div>
		{/each}
	{/if}

	<!-- Orb: hidden when over a button -->
	<div
		class="orb-pos"
		class:orb-hidden={isOverButton || isOverEditor}
		style="left: {cursorX}px; top: {cursorY}px;"
	>
		<div class="orb" class:moving={isMoving}></div>
	</div>
{/if}

{@render children()}

<style>
	/* ── Hide native cursor ── */
	:global(body) {
		cursor: none;
	}
	:global(a, button, input, textarea, select, [role="button"]) {
		cursor: none;
	}

	/* ── Editor area: restore native text cursor ── */
	:global([data-cursor-text]),
	:global([data-cursor-text] *) {
		cursor: text !important;
	}

	/* ── Position wrapper: tracks cursor with no lag ── */
	.orb-pos {
		position: fixed;
		pointer-events: none;
		z-index: 99999;
		transform: translate(-50%, -50%);
		opacity: 1;
		transition: opacity 1s ease;
	}
	.orb-pos.orb-hidden {
		opacity: 0;
	}

	/* ── Button / input hover: sphere-breathe glow ── */
	:global(button:hover),
	:global(input:hover),
	:global(a[role="button"]:hover) {
		transition: box-shadow 0.2s ease !important;
		animation: button-sphere-breathe 2.4s ease-in-out infinite !important;
	}

	/* Cascade title and any data-cursor-glow element — pill-shaped glow */
	:global([data-cursor-glow]:hover) {
		border-radius: 999px !important;
		padding: 4px 12px !important;
		transition:
			box-shadow 0.2s ease,
			padding 0.15s ease !important;
		animation: button-sphere-breathe 2.4s ease-in-out infinite !important;
	}

	@keyframes button-sphere-breathe {
		0%,
		100% {
			box-shadow:
				0 0 8px 3px rgba(236, 72, 153, 0.5),
				0 0 20px 7px rgba(236, 72, 153, 0.22),
				0 0 42px 14px rgba(168, 85, 247, 0.13);
		}
		50% {
			box-shadow:
				0 0 14px 5px rgba(236, 72, 153, 0.72),
				0 0 34px 12px rgba(236, 72, 153, 0.38),
				0 0 65px 22px rgba(168, 85, 247, 0.22);
		}
	}

	/* ── The orb — always a perfect sphere ── */
	.orb {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		pointer-events: none;

		background: radial-gradient(
			circle at 38% 32%,
			rgba(255, 255, 255, 0.92) 0%,
			rgba(249, 168, 212, 0.82) 22%,
			rgba(236, 72, 153, 0.68) 52%,
			rgba(168, 85, 247, 0.45) 78%,
			rgba(109, 40, 217, 0.2) 100%
		);

		box-shadow:
			0 0 5px 2px rgba(236, 72, 153, 0.55),
			0 0 12px 4px rgba(236, 72, 153, 0.22),
			0 0 22px 7px rgba(168, 85, 247, 0.12),
			inset 0 1px 2px rgba(255, 255, 255, 0.65);

		transition: box-shadow 0.18s ease;
		animation: sphere-breathe 2.4s ease-in-out infinite;
	}

	/* Moving: faster breathe, sphere stays round */
	.orb.moving {
		animation: sphere-breathe-fast 0.5s ease-in-out infinite;
	}

	/* ── Tentacle dots ── */
	.jelly-dot {
		position: fixed;
		border-radius: 50%;
		pointer-events: none;
		z-index: 99998;
		transform: translate(-50%, -50%);

		background: radial-gradient(
			circle,
			rgba(var(--c), 0.95) 0%,
			rgba(var(--c), 0.55) 45%,
			rgba(var(--c), 0.1) 75%,
			transparent 100%
		);
		box-shadow: 0 0 5px 2px rgba(var(--c), 0.3);

		animation: dot-fade linear forwards;
	}

	/* ── Keyframes ── */

	/* Idle sphere breathes gently */
	@keyframes sphere-breathe {
		0%,
		100% {
			transform: scale(1);
			box-shadow:
				0 0 6px 2px rgba(236, 72, 153, 0.55),
				0 0 16px 5px rgba(236, 72, 153, 0.22),
				0 0 30px 9px rgba(168, 85, 247, 0.12),
				inset 0 1px 3px rgba(255, 255, 255, 0.65);
		}
		50% {
			transform: scale(1.1);
			box-shadow:
				0 0 10px 3px rgba(236, 72, 153, 0.7),
				0 0 22px 7px rgba(236, 72, 153, 0.32),
				0 0 42px 13px rgba(168, 85, 247, 0.18),
				inset 0 1px 3px rgba(255, 255, 255, 0.65);
		}
	}

	/* Fast breathe while cursor is in motion */
	@keyframes sphere-breathe-fast {
		0%,
		100% {
			transform: scale(1);
			box-shadow:
				0 0 5px 2px rgba(236, 72, 153, 0.6),
				0 0 14px 5px rgba(236, 72, 153, 0.28),
				0 0 26px 8px rgba(168, 85, 247, 0.15),
				inset 0 1px 2px rgba(255, 255, 255, 0.65);
		}
		50% {
			transform: scale(1.18);
			box-shadow:
				0 0 9px 4px rgba(236, 72, 153, 0.8),
				0 0 22px 8px rgba(236, 72, 153, 0.42),
				0 0 40px 14px rgba(168, 85, 247, 0.22),
				inset 0 1px 2px rgba(255, 255, 255, 0.65);
		}
	}

	/* Dot: fade out and shrink */
	@keyframes dot-fade {
		0% {
			opacity: 0.9;
			transform: translate(-50%, -50%) scale(1);
		}
		40% {
			opacity: 0.6;
			transform: translate(-50%, -50%) scale(0.75);
		}
		100% {
			opacity: 0;
			transform: translate(-50%, -50%) scale(0.08);
		}
	}
</style>
