import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import type { PostFeedItem } from "../api";
import { CreatePostForm } from "../components/CreatePostForm";

export const HomePage = () => {
    const [posts, setPosts] = useState<PostFeedItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const loadPosts = async () => {
        try {
            setLoading(true);
            setError("");

            const response = await api.getPosts();

            setPosts(response.data);
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Failed to load posts."
            );
        } finally {
            setLoading(false);
        }
    };
    useEffect(() => {
        loadPosts();
    }, []);

    return (
        <div>
            <section className="hero">
                <h1>GopherSocial</h1>
                <p>A social network for gophers.</p>
            </section>

            <CreatePostForm onCreated={loadPosts} />

            <section className="feed">
                <h2>Latest posts</h2>

                {loading && (
                    <div className="loading">
                        Loading posts...
                    </div>
                )}

                {error && (
                    <div className="error">
                        {error}
                    </div>
                )}

                {!loading && !error && posts.length === 0 && (
                    <div className="card empty-feed">
                        <p>No posts yet.</p>
                    </div>
                )}

                <div className="post-list">
                    {posts.map((item) => {

                        return (
                            <article className="card feed-post" key={item.id}>
                                <div className="post-author">
                                    <Link to={`/users/${item.user_summary.id}`}>
                                        {item.user_summary.username}
                                    </Link>

                                    <span>
                                        {new Date(item.created_at).toLocaleString()}
                                    </span>
                                </div>

                                <Link to={`/posts/${item.id}`}>
                                    <h2>{item.title}</h2>
                                </Link>

                                <p className="post-content">{item.content}</p>

                                {item.tags?.length > 0 && (
                                    <div className="tags">
                                        {item.tags.map((tag) => (
                                            <span key={tag}>#{tag}</span>
                                        ))}
                                    </div>
                                )}

                                <div className="post-meta">
                                    {item.comments_count}{" "}
                                    {item.comments_count === 1 ? "comment" : "comments"}
                                </div>
                            </article>
                        );
                    })}

                </div>
            </section>
        </div>
    );
};