import { useEffect, useMemo, useRef, useState } from "react";

type Props = {
    peaks: number[];              // values 0..1000
    durationMs?: number | null;
    audioRef: React.RefObject<HTMLAudioElement | null>;
    height?: number;
};

function clamp01(x: number) {
    return Math.max(0, Math.min(1, x));
}

export default function WaveformCanvas({ peaks, durationMs, audioRef, height = 80 }: Props) {
    const canvasRef = useRef<HTMLCanvasElement | null>(null);
    const rafRef = useRef<number | null>(null);
    const draggingRef = useRef(false);

    const [hoverP, setHoverP] = useState<number | null>(null);

    const durS = useMemo(() => (durationMs ? durationMs / 1000 : 0), [durationMs]);

    // Draw waveform + playhead
    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const ctx = canvas.getContext("2d");
        if (!ctx) return;

        // Handle HiDPI
        const dpr = window.devicePixelRatio || 1;
        const cssW = canvas.clientWidth;
        const cssH = height;
        canvas.width = Math.floor(cssW * dpr);
        canvas.height = Math.floor(cssH * dpr);
        canvas.style.height = `${cssH}px`;
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

        const draw = () => {
            const w = canvas.clientWidth;
            const h = cssH;

            // Clear
            ctx.clearRect(0, 0, w, h);

            // Background
            ctx.fillStyle = "#fff";
            ctx.fillRect(0, 0, w, h);

            // Waveform bars
            const N = peaks.length;
            if (N > 0) {
                const mid = h / 2;
                const step = w / N;

                ctx.fillStyle = "#111"; // waveform color
                for (let i = 0; i < N; i++) {
                    const v = peaks[i] ?? 0;
                    const amp = (v / 1000) * (h * 0.48);
                    const x = i * step;

                    // draw symmetric bar around center
                    const barW = Math.max(1, step); // if step < 1, still draw
                    ctx.fillRect(x, mid - amp, barW, amp * 2);
                }
            } else {
                ctx.fillStyle = "#666";
                ctx.font = "12px system-ui, -apple-system, sans-serif";
                ctx.fillText("No waveform data yet", 10, 20);
            }

            // Hover line
            if (hoverP != null) {
                ctx.strokeStyle = "rgba(0,0,0,0.2)";
                ctx.beginPath();
                ctx.moveTo(hoverP * w, 0);
                ctx.lineTo(hoverP * w, h);
                ctx.stroke();
            }

            // Playhead
            const a = audioRef.current;
            if (a && durS > 0) {
                const p = clamp01(a.currentTime / durS);
                ctx.strokeStyle = "rgba(255,0,0,0.8)";
                ctx.lineWidth = 2;
                ctx.beginPath();
                ctx.moveTo(p * w, 0);
                ctx.lineTo(p * w, h);
                ctx.stroke();
            }

            rafRef.current = requestAnimationFrame(draw);
        };

        rafRef.current = requestAnimationFrame(draw);
        return () => {
            if (rafRef.current != null) cancelAnimationFrame(rafRef.current);
            rafRef.current = null;
        };
    }, [peaks, height, durS, audioRef, hoverP]);

    // Mouse interactions
    function seekFromEvent(clientX: number) {
        const canvas = canvasRef.current;
        const a = audioRef.current;
        if (!canvas || !a || durS <= 0) return;

        const rect = canvas.getBoundingClientRect();
        const p = clamp01((clientX - rect.left) / rect.width);
        a.currentTime = p * durS;
    }

    return (
        <div style={{ marginTop: 10 }}>
            <div style={{ border: "1px solid #eee", borderRadius: 12, overflow: "hidden" }}>
                <canvas
                    ref={canvasRef}
                    style={{ width: "100%", height, display: "block", cursor: "pointer" }}
                    onMouseMove={(e) => {
                        const canvas = canvasRef.current;
                        if (!canvas) return;
                        const rect = canvas.getBoundingClientRect();
                        const p = clamp01((e.clientX - rect.left) / rect.width);
                        setHoverP(p);
                        if (draggingRef.current) seekFromEvent(e.clientX);
                    }}
                    onMouseLeave={() => {
                        setHoverP(null);
                        draggingRef.current = false;
                    }}
                    onMouseDown={(e) => {
                        draggingRef.current = true;
                        seekFromEvent(e.clientX);
                    }}
                    onMouseUp={() => {
                        draggingRef.current = false;
                    }}
                    onClick={(e) => {
                        // Click seeks as well
                        seekFromEvent(e.clientX);
                    }}
                />
            </div>

            {/* Hover time hint */}
            {hoverP != null && durS > 0 && (
                <div style={{ marginTop: 6, fontSize: 12, color: "#666" }}>
                    Hover: {Math.round(hoverP * durS)}s
                </div>
            )}
        </div>
    );
}
