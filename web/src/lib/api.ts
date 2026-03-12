export const API_BASE = "http://localhost:8080";

export type Track = {
    id: number;
    title: string;
    audio_name?: string | null;
    audio_mime?: string | null;
    duration_ms?: number | null;
    waveform_peaks?: number[];
    created_at: string;
};

export type Comment = {
    id: number;
    track_id: number;
    author: string;
    timestamp_ms: number;
    text: string;
    created_at: string;
};

export async function listTracks(): Promise<Track[]> {
    const res = await fetch(`${API_BASE}/tracks`);
    if (!res.ok) throw new Error(`listTracks ${res.status}`);
    return res.json();
}

export async function getTrack(id: number): Promise<Track> {
    const res = await fetch(`${API_BASE}/tracks/${id}`);
    if (!res.ok) throw new Error(`getTrack ${res.status}`);
    return res.json();
}

export async function listComments(trackId: number): Promise<Comment[]> {
    const res = await fetch(`${API_BASE}/tracks/${trackId}/comments`);
    if (!res.ok) throw new Error(`listComments ${res.status}`);
    return res.json();
}

export async function createComment(trackId: number, payload: { author: string; text: string; timestamp_ms: number }): Promise<Comment> {
    const res = await fetch(`${API_BASE}/tracks/${trackId}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
    });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
}

export async function uploadTrack(title: string, file: File): Promise<Track> {
    const form = new FormData();
    form.append("title", title);
    form.append("audio", file);

    const res = await fetch(`${API_BASE}/tracks`, {
        method: "POST",
        body: form,
    });

    if (!res.ok) throw new Error(await res.text());
    return res.json();
}

export function uploadTrackWithProgress(
    title: string,
    file: File,
    onProgress: (percent: number) => void
): Promise<Track> {
    return new Promise((resolve, reject) => {
        const form = new FormData();
        form.append("title", title);
        form.append("audio", file);

        const xhr = new XMLHttpRequest();
        xhr.open("POST", `${API_BASE}/tracks`);

        xhr.upload.onprogress = (e) => {
            if (!e.lengthComputable) return;
            const percent = Math.round((e.loaded / e.total) * 100);
            onProgress(percent);
        };

        xhr.onload = () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                try {
                    resolve(JSON.parse(xhr.responseText));
                } catch (err) {
                    reject(err);
                }
            } else {
                reject(new Error(xhr.responseText || `Upload failed: ${xhr.status}`));
            }
        };

        xhr.onerror = () => reject(new Error("Network error during upload"));
        xhr.send(form);
    });
}

export async function createShare(trackId: number): Promise<{ token: string; url: string }> {
    const res = await fetch(`${API_BASE}/tracks/${trackId}/share`, { method: "POST" });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
}

export async function resolveShare(token: string): Promise<{ token: string; track_id: number; expires_at?: string | null }> {
    const res = await fetch(`${API_BASE}/api/share/${token}`);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
}
