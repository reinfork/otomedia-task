import axios from "axios";

const baseURL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:8080";

export const api = axios.create({
  baseURL,
  timeout: 10000,
  headers: { "Content-Type": "application/json" },
});

export type TaskStatus = "todo" | "in_progress" | "done";

export interface Task {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  assignee: string;
  created_at: string;
  updated_at: string;
}

export interface TaskListParams {
  status?: string;
  keyword?: string;
  assignee?: string;
  page?: number;
  limit?: number;
  sort?: string;
}

export interface TaskListResponse {
  data: Task[];
  meta: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export async function fetchTasks(params: TaskListParams): Promise<TaskListResponse> {
  const res = await api.get("/api/tasks", { params });
  return res.data;
}

export async function updateTask(id: string, payload: Partial<Task>): Promise<Task> {
  const res = await api.put(`/api/tasks/${id}`, payload);
  return res.data.data;
}

export async function deleteTask(id: string): Promise<void> {
  await api.delete(`/api/tasks/${id}`);
}
