import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listTracks, uploadTrackWithProgress } from "../lib/api";
import type { Track } from "../lib/api";

export default function TracksPage() {
    const [items, setItems] = useState<Track[]>([]);
    const [err, setErr] = useState("");

    const [title, setTitle] = useState("New demo");
    const [file, setFile] = useState<File | null>(null);
    const [busy, setBusy] = useState(false);
    const [progress, setProgress] = useState<number>(0);

    async function onUpload() {
        if (!file) {
            setErr("Choose an audio file first");
            return;
        }

        setBusy(true);
        setErr("");
        setProgress(0);

        try {
            const t = await uploadTrackWithProgress(title, file, setProgress);

            // optional: refresh list
            const refreshed = await listTracks();
            setItems(refreshed);

            // go to track page immediately; it will show "Analyzing…" until SSE track.analyzed arrives
            window.location.href = `/tracks/${t.id}`;
        } catch (e) {
            setErr(String(e));
        } finally {
            setBusy(false);
        }
    }

    useEffect(() => {
        listTracks().then(setItems).catch((e) => setErr(String(e)));
    }, []);

    return (
        <div>
            <h3 style={{ marginTop: 0 }}>Tracks</h3>
            <div style={{ border: "1px solid #eee", borderRadius: 14, padding: 12, marginBottom: 14 }}>
                <div style={{ fontWeight: 600, marginBottom: 8 }}>Upload</div>
                <div style={{ display: "flex", gap: 8, flexWrap: "wrap", alignItems: "center" }}>
                    <input
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="title"
                        style={{ padding: "10px 12px", borderRadius: 12, border: "1px solid #ddd" }}
                    />
                    <input
                        type="file"
                        accept="audio/*"
                        onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                    />
                    <button
                        onClick={onUpload}
                        disabled={busy}
                        style={{ padding: "10px 12px", borderRadius: 12, border: "1px solid #ddd", cursor: "pointer" }}
                    >
                        {busy ? "Uploading…" : "Upload"}
                    </button>
                </div>
            </div>
            {err && <div style={{ color: "crimson" }}>{err}</div>}
            {busy && (
                <div style={{ marginTop: 10 }}>
                    <div style={{ fontSize: 12, color: "#666", marginBottom: 6 }}>
                        Uploading… {progress}%
                    </div>
                    <div style={{ height: 10, borderRadius: 999, background: "#eee", overflow: "hidden" }}>
                        <div style={{ height: "100%", width: `${progress}%`, background: "#111" }} />
                    </div>
                </div>
            )}

            <div style={{ display: "grid", gap: 10 }}>
                {items.map((t) => (
                    <Link key={t.id} to={`/tracks/${t.id}`} style={{ textDecoration: "none", color: "inherit" }}>
                        <div style={{ border: "1px solid #eee", borderRadius: 14, padding: 12 }}>
                            <div style={{ fontWeight: 600 }}>{t.title}</div>
                            <div style={{ color: "#666", fontSize: 12 }}>
                                id={t.id} • {t.duration_ms ? `${Math.round(t.duration_ms / 1000)}s` : "analyzing…"} • {t.audio_name ?? "no file"}
                            </div>
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    );
}