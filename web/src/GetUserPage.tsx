import { useState } from "react"
import { useParams } from "react-router-dom"
import { API_URL } from "./App"

// 1. Define the user model
interface User {
    created_at: string
    email: string
    id: number
    is_active: boolean
    role_id: number
    username: string
}

// 2. Define the outer API response structure
interface ApiResponse {
    data: User
}

export const GetUserPage = () => {
    const { userID = '' } = useParams()
    const [userData, setUserData] = useState<User | null>(null)

    const handleConfirm = async () => {
        try {
            const response = await fetch(`${API_URL}/api/v1/users/${userID}`, {
                method: "GET"
            })

            // 1. Read the response as raw text first
            const textResponse = await response.text()
            console.log("Raw server response:", textResponse)

            if (response.ok) {
                // 2. Safely parse the text as JSON
                const json: ApiResponse = JSON.parse(textResponse)
                setUserData(json.data)
            } else {
                alert(`Failed to get user. Status: ${response.status}. Check console for details.`)
            }
        } catch (error) {
            console.error("Error fetching user:", error)
            alert("Something went wrong. Check the console.")
        }
    }

    return (
        <div>
            <h1>User Details</h1>
            <button onClick={handleConfirm}>Click to get</button>

            {userData ? (
                <div style={{ marginTop: "20px" }}>
                    <h2>User Data:</h2>
                    <p><strong>ID:</strong> {userData.id}</p>
                    <p><strong>Username:</strong> {userData.username}</p>
                    <p><strong>Email:</strong> {userData.email}</p>
                    <p><strong>Active:</strong> {userData.is_active ? "Yes" : "No"}</p>
                    <p><strong>Created At:</strong> {userData.created_at}</p>
                </div>
            ) : (
                <p style={{ marginTop: "20px" }}>Click the button to load user data.</p>
            )}
        </div>
    )
}