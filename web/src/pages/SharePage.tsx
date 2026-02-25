import { useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { API_BASE, createComment, getTrack, listComments, resolveShare } from "../lib/api";
import type { Comment, Track } from "../lib/api";

function msToTime(ms: number) {
    const s = Math.floor(ms / 1000);
    const m = Math.floor(s / 60);
    const ss = String(s % 60).padStart(2, "0");
    return `${m}:${ss}`;
}

export default function SharePage() {
    const { token = "" } = useParams();

    const audioRef = useRef<HTMLAudioElement | null>(null);

    const [trackId, setTrackId] = useState<number | null>(null);
    const [track, setTrack] = useState<Track | null>(null);
    const [comments, setComments] = useState<Comment[]>([]);
    const [author, setAuthor] = useState("guest");
    const [text, setText] = useState("");
    const [err, setErr] = useState("");

    const audioUrl = useMemo(() => (trackId ? `${API_BASE}/tracks/${trackId}/audio` : ""), [trackId]);

    useEffect(() => {
        (async () => {
            try {
                setErr("");
                const resolved = await resolveShare(token);
                setTrackId(resolved.track_id);
            } catch (e) {
                setErr(String(e));
            }
        })();
    }, [token]);

    useEffect(() => {
        if (!trackId) return;
        setErr("");
        getTrack(trackId).then(setTrack).catch((e) => setErr(String(e)));
        listComments(trackId).then(setComments).catch((e) => setErr(String(e)));
    }, [trackId]);

    // SSE subscribe
    useEffect(() => {
        if (!trackId) return;
        const es = new EventSource(`${API_BASE}/tracks/${trackId}/events`);
        es.onmessage = (e) => {
            try {
                const msg = JSON.parse(e.data);
                if (msg.type === "comment.created") {
                    setComments((prev) => {
                        const c = msg.data as Comment;
                        if (prev.some((x) => x.id === c.id)) return prev;
                        return [...prev, c];
                    });
                }

                if (msg.type === "track.analyzed") {
                    getTrack(trackId).then(setTrack).catch(() => {});
                }
            } catch {}
        };
        return () => es.close();
    }, [trackId]);

    async function onSend() {
        if (!trackId) return;
        try {
            const t = audioRef.current?.currentTime ?? 0;
            const timestamp_ms = Math.floor(t * 1000);
            const c = await createComment(trackId, { author, text, timestamp_ms });
            // Optimistic de-dupe
            setComments((prev) => (prev.some((x) => x.id === c.id) ? prev : [...prev, c]));
            setText("");
        } catch (e) {
            setErr(String(e));
        }
    }

    function jump(ms: number) {
        const a = audioRef.current;
        if (!a) return;
        a.currentTime = ms / 1000.0;
        a.play().catch(() => {});
    }

    return (
        <div>
            <h3 style={{ marginTop: 0 }}>{track ? track.title : "Shared track"}</h3>
            {err && <div style={{ color: "crimson", marginBottom: 10 }}>{err}</div>}

            {trackId && <audio ref={audioRef} controls style={{ width: "100%" }} src={audioUrl} />}

            <h4 style={{ marginTop: 18 }}>Add comment</h4>
            <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                <input value={author} onChange={(e) => setAuthor(e.target.value)} placeholder="author" style={{ padding: "10px 12px", borderRadius: 12, border: "1px solid #ddd" }} />
                <input value={text} onChange={(e) => setText(e.target.value)} placeholder="comment" style={{ padding: "10px 12px", borderRadius: 12, border: "1px solid #ddd", minWidth: 320 }} />
                <button onClick={onSend} style={{ padding: "10px 12px", borderRadius: 12, border: "1px solid #ddd", cursor: "pointer" }}>
                    Send at current time
                </button>
            </div>

            <h4 style={{ marginTop: 18 }}>Comments</h4>
            <div style={{ display: "grid", gap: 10 }}>
                {comments.map((c) => (
                    <div key={c.id} onClick={() => jump(c.timestamp_ms)} style={{ borderBottom: "1px solid #eee", paddingBottom: 10, cursor: "pointer" }}>
                        <div style={{ color: "#666", fontSize: 12 }}>
                            {msToTime(c.timestamp_ms)} — {c.author}
                        </div>
                        <div>{c.text}</div>
                    </div>
                ))}
            </div>
        </div>
    );
}