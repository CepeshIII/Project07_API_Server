import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api";
import type { Follower } from "../api";
import type { User } from "../api";

export const GetUserPage = () => {
    const { userID = "" } = useParams();

    const [userData, setUserData] = useState<User | null>(null);
    const [followers, setFollowers] = useState<Follower[]>([]);

    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    const id = Number(userID);

    const loadUser = async () => {
        try {
            setLoading(true);
            setError("");

            const [userResponse, followersResponse] =
                await Promise.all([
                    api.getUser(id),
                    api.followers(id),
                ]);

            setUserData(userResponse.data);
            setFollowers(followersResponse.data);
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Failed to load user."
            );
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (!Number.isNaN(id)) {
            loadUser();
        }
    }, [id]);

    if (loading) {
        return <div className="loading">Loading user...</div>;
    }

    if (error) {
        return <div className="error-page">{error}</div>;
    }

    if (!userData) {
        return <div>User not found.</div>;
    }

    return (
        <div className="profile-page">
            <section className="card profile">
                <div className="avatar">
                    {userData.username.charAt(0).toUpperCase()}
                </div>

                <h1>{userData.username}</h1>

                <p className="muted">{userData.email}</p>

                <div className="profile-info">
                    <div>
                        <strong>ID</strong>
                        <span>{userData.id}</span>
                    </div>

                    <div>
                        <strong>Status</strong>
                        <span>
                            {userData.is_active
                                ? "Active"
                                : "Inactive"}
                        </span>
                    </div>

                    <div>
                        <strong>Role</strong>
                        <span>{userData.role_id}</span>
                    </div>

                    <div>
                        <strong>Created</strong>
                        <span>
                            {new Date(
                                userData.created_at
                            ).toLocaleDateString()}
                        </span>
                    </div>
                </div>

                <div className="profile-actions">
                    <button
                        onClick={async () => {
                            try {
                                await api.follow(userData.id);
                                alert("Followed");
                            } catch (error) {
                                alert(
                                    error instanceof Error
                                        ? error.message
                                        : "Failed to follow."
                                );
                                return;
                            } finally {
                                setLoading(false);
                            }
                        }}
                    >
                        Follow
                    </button>

                    <button
                        className="secondary"
                        onClick={async () => {
                            try {
                                await api.unfollow(userData.id);
                                alert("Unfollowed");
                            } catch (error) {
                                alert(
                                    error instanceof Error
                                        ? error.message
                                        : "Failed to unfollow."
                                );

                                return;
                            } finally {
                                setLoading(false);
                            }

                        }}
                    >
                        Unfollow
                    </button>
                </div>
            </section>

            <section className="card">
                <h2>Followers</h2>

                {followers === null || followers.length === 0 ? (
                    <p className="muted">
                        This user has no followers.
                    </p>
                ) : (
                    <div className="followers">
                        {followers.map((follower) => (
                            <Link
                                to={`/users/${follower.user_summary.id}`}
                            >
                                {follower.user_summary.username}
                            </Link>
                        ))}
                    </div>
                )}
            </section>
        </div>
    );
};