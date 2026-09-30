import { API_URL } from "./App";

export interface User {
    id: number;
    username: string;
    email: string;
    created_at: string;
    is_active: boolean;
    role_id: number;
}

export interface Post {
    id: number;
    title: string;
    content: string;
    tags: string[];
    created_at: string;
    updated_at: string;
    user_id: number;
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
export interface PostWithMetadata {
    post_data: Post;
    comments_count: number;
}
export interface PostWithComments {
    post_data: Post;
    post_comments: CommentWithUser[];
}

export interface Follower {
    created_at: string;
    follower_id: number;
    user_id: number;
}

export interface ApiResponse<T> {
    data: T;
}

export interface ApiError {
    error: string;
}

async function request<T>(
    path: string,
    options: RequestInit = {}
): Promise<T> {
    const response = await fetch(`${API_URL}${path}`, {
        ...options,
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
            ...options.headers,
        },
    });

    const text = await response.text();

    let data: unknown = null;

    if (text) {
        try {
            data = JSON.parse(text);
        } catch {
            data = text;
        }
    }

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
        request<ApiResponse<PostWithMetadata[]>>("/posts"),
};