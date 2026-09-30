import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api";

export const ConfirmationPage = () => {
    const { token = "" } = useParams();
    const navigate = useNavigate();

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const handleConfirm = async () => {
        try {
            setLoading(true);
            setError("");

            await api.activate(token);

            navigate("/login");
        } catch (error) {
            setError(
                error instanceof Error
                    ? error.message
                    : "Failed to activate account."
            );
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="center-page">
            <div className="card confirmation">
                <h1>Confirm account</h1>

                <p>
                    Click the button below to activate your account.
                </p>

                {error && <div className="error">{error}</div>}

                <button
                    onClick={handleConfirm}
                    disabled={loading}
                >
                    {loading ? "Activating..." : "Activate account"}
                </button>
            </div>
        </div>
    );
};