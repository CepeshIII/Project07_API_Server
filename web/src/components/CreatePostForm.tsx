import type { FormEvent } from "react";
import { useState } from "react";
import { api } from "../api";

interface Props {
    onCreated?: () => void;
}

export const CreatePostForm = ({ onCreated }: Props) => {
    const [title, setTitle] = useState("");
    const [content, setContent] = useState("");
    const [tags, setTags] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const handleSubmit = async (event: FormEvent) => {
        event.preventDefault();

        if (!title.trim() || !content.trim()) {
            setError("Title and content are required.");
            return;
        }

        try {
            setLoading(true);
            setError("");

            await api.createPost({
                title,
                content,
                tags: tags
                    .split(",")
                    .map((tag) => tag.trim())
                    .filter(Boolean),
            });

            setTitle("");
            setContent("");
            setTags("");

            onCreated?.();
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Failed to create post."
            );
        } finally {
            setLoading(false);
        }
    };

    return (
        <form className="card post-form" onSubmit={handleSubmit}>
            <h2>Create post</h2>

            <input
                placeholder="Title..."
                value={title}
                maxLength={200}
                onChange={(e) => setTitle(e.target.value)}
            />

            <textarea
                placeholder="What's in your mind?"
                value={content}
                maxLength={1000}
                rows={5}
                onChange={(e) => setContent(e.target.value)}
            />

            <input
                placeholder="Tags: go, programming, backend"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
            />

            {error && <div className="error">{error}</div>}

            <button disabled={loading}>
                {loading ? "Publishing..." : "Share"}
            </button>
        </form>
    );
};