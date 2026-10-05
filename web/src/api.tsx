import { API_URL } from "./App";

// ==========================================
// 1. ALL YOUR INTERFACES (Unchanged)
// ==========================================

export interface User {
    id: number;
    username: string;
    email: string;
    created_at: string;
    is_active: boolean;
    role_id: number;
}

export interface UserSummary {
    id: number;
    username: string;
}

export interface Post {
    id: number;
    title: string;
    content: string;
    tags: string[];
    created_at: string;
    updated_at: string;
    version: number;
    user: User;
}

export interface Comment {
    id: number;
    post_id: number;
    user_id: number;
    content: string;
    created_at: string;
}

export interface CommentWithUser extends Comment {
    username: string;
}

// export interface PostWithMetadata {
//     post_data: Post;
//     comments_count: number;
// }

export interface PostWithComments {
    post_data: Post;
    post_comments: CommentWithUser[];
}

export interface PostFeedItem {
    id: number
    title: string
    content: string
    tags: string[]
    created_at: string
    version: number
    user_summary: UserSummary
    comments_count: number
}

export interface Follower {
    user_summary: UserSummary;
    created_at: string
}

export interface ApiResponse<T> {
    data: T;
}

export interface ApiError {
    error: string;
}

// ==========================================
// 2. CONCURRENT REFRESH QUEUE LOGIC
// ==========================================
// Since multiple requests might fail at the exact same time with a 401,
// we use a queue so we only call /auth/refresh ONCE, and make the others wait.

let isRefreshing = false;
let failedQueue: Array<{
    resolve: (value?: any) => void; // <-- Updated to accept optional value
    reject: (error: any) => void;
}> = [];

const processQueue = (error: any) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve();
        }
    });
    failedQueue = [];
};

// ==========================================
// 3. ENHANCED FETCH REQUEST FUNCTION
// ==========================================
/**
 * This wrapper function replaces Axios. Every time you make a request,
 * it runs through here. If it gets a 401 error, it automatically attempts 
 * to refresh the token and retry the request.
 */
async function request<T>(
    path: string,
    options: RequestInit = {},
    isRetry = false // Prevents infinite loops if the refresh itself fails
): Promise<T> {
    // 1. Send the fetch request
    const response = await fetch(`${API_URL}${path}`, {
        ...options,
        credentials: "include", // Crucial for sending/receiving HttpOnly cookies
        headers: {
            "Content-Type": "application/json",
            ...options.headers,
        },
    });

    // 2. Read the response text safely
    const text = await response.text();
    let data: unknown = null;

    if (text) {
        try {
            data = JSON.parse(text);
        } catch {
            data = text;
        }
    }

    // 3. HANDLE 401 UNAUTHORIZED (Token Expired)
    if (response.status === 401 && !isRetry) {
        // If we are already trying to refresh, queue this request until it finishes
        if (isRefreshing) {
            return new Promise((resolve, reject) => {
                failedQueue.push({ resolve, reject });
            }).then(() => {
                // Once the refresh queue resolves, retry the original request
                return request<T>(path, options, true);
            });
        }

        isRefreshing = true;

        try {
            // Call your backend refresh endpoint to issue a new access token cookie
            const refreshResponse = await fetch(`${API_URL}/auth/refresh`, {
                method: "POST",
                credentials: "include",
            });

            if (!refreshResponse.ok) {
                throw new Error("Refresh token expired");
            }

            // Success! Wake up everyone waiting in the queue
            isRefreshing = false;
            processQueue(null);

            // Retry the original request that failed with 401
            return request<T>(path, options, true);
        } catch (refreshError) {
            // Refresh token failed completely -> Kick user out to login
            isRefreshing = false;
            processQueue(refreshError);
            window.location.href = "/login";
            throw refreshError;
        }
    }

    // 4. Handle other general HTTP errors
    if (!response.ok) {
        const error =
            typeof data === "object" &&
                data !== null &&
                "error" in data
                ? String((data as ApiError).error)
                : `Request failed with status ${response.status}`;

        throw new Error(error);
    }

    return data as T;
}

// ==========================================
// 4. API OBJECT METHODS (Unchanged usage)
// ==========================================

export const api = {
    health: () =>
        request<Record<string, string>>("/health"),

    register: (payload: {
        username: string;
        email: string;
        password: string;
    }) =>
        request<ApiResponse<{
            user: User;
            verification_token: string;
        }>>("/auth/register", {
            method: "POST",
            body: JSON.stringify(payload),
        }),

    login: (payload: {
        username: string;
        password: string;
    }) =>
        request<ApiResponse<{
            user: User;
            verification_token?: string;
        }>>("/auth/login", {
            method: "POST",
            body: JSON.stringify(payload),
        }),

    refresh: () =>
        request<ApiResponse<string>>("/auth/refresh", {
            method: "POST",
        }),

    activate: (token: string) =>
        request<ApiResponse<string>>(`/users/activate/${token}`, {
            method: "PUT",
        }),

    getUser: (id: number) =>
        request<ApiResponse<User>>(`/users/${id}`),

    deleteUser: (id: number) =>
        request<ApiResponse<string>>(`/users/${id}`, {
            method: "DELETE",
        }),

    follow: (id: number) =>
        request<ApiResponse<string>>(`/users/${id}/follow`, {
            method: "PUT",
        }),

    unfollow: (id: number) =>
        request<ApiResponse<string>>(`/users/${id}/unfollow`, {
            method: "PUT",
        }),

    followers: (id: number) =>
        request<ApiResponse<Follower[]>>(`/users/${id}/followers`),

    createPost: (payload: {
        title: string;
        content: string;
        tags?: string[];
    }) =>
        request<Post>("/posts/", {
            method: "POST",
            body: JSON.stringify(payload),
        }),

    getPost: (id: number) =>
        request<ApiResponse<PostWithComments>>(`/posts/${id}`),

    updatePost: (
        id: number,
        payload: {
            title?: string;
            content?: string;
            tags?: string[];
        }
    ) =>
        request<Post>(`/posts/${id}`, {
            method: "PATCH",
            body: JSON.stringify(payload),
        }),

    deletePost: (id: number) =>
        request<ApiResponse<string>>(`/posts/${id}`, {
            method: "DELETE",
        }),

    getComments: (postId: number) =>
        request<ApiResponse<Comment[]>>(
            `/posts/${postId}/comments`
        ),

    createComment: (postId: number, content: string) =>
        request<ApiResponse<string>>(`/posts/${postId}/comments`, {
            method: "POST",
            body: JSON.stringify({ content }),
        }),

    getPosts: () =>
        request<ApiResponse<PostFeedItem[]>>("/posts"),
};