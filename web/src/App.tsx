import { Link, Route, Routes } from "react-router-dom";
import TracksPage from "./pages/TracksPage";
import TrackPage from "./pages/TrackPage";

export default function App() {
    return (
        <div style={{ maxWidth: 920, margin: "24px auto", padding: "0 14px", fontFamily: "system-ui, -apple-system, sans-serif" }}>
            <header style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", gap: 12 }}>
                <h2 style={{ margin: 0 }}>
                    <Link to="/" style={{ textDecoration: "none", color: "inherit" }}>MusicReview</Link>
                </h2>
                <span style={{ color: "#666" }}>Go + Postgres + SSE</span>
            </header>
            <hr />
            <Routes>
                <Route path="/" element={<TracksPage />} />
                <Route path="/tracks/:id" element={<TrackPage />} />
            </Routes>
        </div>
    );
}