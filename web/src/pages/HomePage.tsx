import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api";
import type { Post } from "../api";
import type { PostWithMetadata } from "../api";
import { CreatePostForm } from "../components/CreatePostForm";

export const HomePage = () => {
    const [posts, setPosts] = useState<PostWithMetadata[]>([]);
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
                        const post = item.post_data;

                        return (
                            <article className="card feed-post" key={post.id}>
                                <div className="post-author">
                                    <Link to={`/users/${post.user.id}`}>
                                        {post.user.username}
                                    </Link>

                                    <span>
                                        {new Date(post.created_at).toLocaleString()}
                                    </span>
                                </div>

                                <Link to={`/posts/${post.id}`}>
                                    <h2>{post.title}</h2>
                                </Link>

                                <p className="post-content">{post.content}</p>

                                {post.tags?.length > 0 && (
                                    <div className="tags">
                                        {post.tags.map((tag) => (
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