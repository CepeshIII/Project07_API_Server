import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api";

export const RegisterPage = () => {
    const navigate = useNavigate();

    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (event: FormEvent) => {
        event.preventDefault();

        try {
            setLoading(true);
            setError("");

            const response = await api.register({
                username,
                email,
                password,
            });

            /*
             * Your backend returns verification_token.
             *
             * In production the token should normally be sent
             * by email and not displayed to the user.
             *
             * For development we can redirect directly.
             */
            navigate(`/activate/${response.data.verification_token}`);
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Registration failed."
            );
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-page">
            <form className="card auth-form" onSubmit={handleSubmit}>
                <h1>Create account</h1>

                <input
                    type="text"
                    placeholder="Username"
                    maxLength={100}
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />

                <input
                    type="email"
                    placeholder="Email"
                    maxLength={255}
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />

                <input
                    type="password"
                    placeholder="Password"
                    minLength={3}
                    maxLength={72}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />

                {error && <div className="error">{error}</div>}

                <button disabled={loading}>
                    {loading ? "Creating..." : "Register"}
                </button>

                <p>
                    Already have an account?{" "}
                    <Link to="/login">Login</Link>
                </p>
            </form>
        </div>
    );
};