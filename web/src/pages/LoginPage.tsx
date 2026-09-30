import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api";

export const LoginPage = () => {
    const navigate = useNavigate();

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (event: FormEvent) => {
        event.preventDefault();

        try {
            setLoading(true);
            setError("");

            await api.login({
                username,
                password,
            });

            navigate("/");
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Login failed."
            );
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-page">
            <form className="card auth-form" onSubmit={handleSubmit}>
                <h1>Login</h1>

                <input
                    type="text"
                    placeholder="Username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />

                <input
                    type="password"
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />

                {error && <div className="error">{error}</div>}

                <button disabled={loading}>
                    {loading ? "Logging in..." : "Login"}
                </button>

                <p>
                    Don't have an account?{" "}
                    <Link to="/register">Register</Link>
                </p>
            </form>
        </div>
    );
};