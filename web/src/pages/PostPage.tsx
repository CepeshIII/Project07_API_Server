import type { FormEvent } from "react";
import { useState } from "react";
import { useEffect } from "react";

import { Link, useParams } from "react-router-dom";
import { api } from "../api";
import type { PostWithComments } from "../api";


export const PostPage = () => {
    const { postID = "" } = useParams();

    const [post, setPost] = useState<PostWithComments | null>(null);
    const [comment, setComment] = useState("");
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    const id = Number(postID);

    const loadPost = async () => {
        try {
            setLoading(true);

            const response = await api.getPost(id);

            setPost(response.data);
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Failed to load post."
            );
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadPost();
    }, [id]);

    const handleComment = async (event: FormEvent) => {
        event.preventDefault();

        if (!comment.trim()) {
            return;
        }

        try {
            await api.createComment(id, comment);

            setComment("");

            await loadPost();
        } catch (error) {
            alert(
                error instanceof Error
                    ? error.message
                    : "Failed to create comment."
            );
        }
    };

    if (loading) {
        return <div className="loading">Loading post...</div>;
    }

    if (error) {
        return <div className="error-page">{error}</div>;
    }

    if (!post) {
        return <div>Post not found.</div>;
    }


    return (
        <div className="post-page">
            <article className="card post">
                <div className="post-author">
                    <Link to={`/users/${post.post_data.user.id}`}>
                        {post.post_data.user.username}
                    </Link>

                    <span>
                        {new Date(
                            post.post_data.created_at
                        ).toLocaleString()}
                    </span>
                </div>

                <h1>{post.post_data.title}</h1>

                <p className="post-content">
                    {post.post_data.content}
                </p>

                {post.post_data.tags?.length > 0 && (
                    <div className="tags">
                        {post.post_data.tags.map((tag) => (
                            <span key={tag}>#{tag}</span>
                        ))}
                    </div>
                )}
            </article>

            <section className="card comments">
                <h2>
                    Comments ({post.post_comments.length})
                </h2>

                <form onSubmit={handleComment}>
                    <textarea
                        placeholder="Write a comment..."
                        value={comment}
                        onChange={(e) =>
                            setComment(e.target.value)
                        }
                        rows={3}
                    />

                    <button>Add comment</button>
                </form>

                <div className="comment-list">
                    {post.post_comments.map((comment) => (
                        <article
                            className="comment"
                            key={comment.id}
                        >
                            <div className="comment-header">
                                <strong>
                                    <Link to={`/users/${comment.user_id}`}>
                                        {comment.username}
                                    </Link>
                                </strong>

                                <span>
                                    {new Date(
                                        comment.created_at
                                    ).toLocaleString()}
                                </span>
                            </div>

                            <p>{comment.content}</p>
                        </article>
                    ))}
                </div>
            </section>
        </div>
    );
};