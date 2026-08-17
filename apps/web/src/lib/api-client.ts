import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios';
import { env } from '#/env';
import type { ApiErrorResponse } from '#/interfaces/api';

export class ApiError extends Error {
  readonly status: number;
  readonly data: unknown;

  constructor(status: number, message: string, data?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
  }
}

export function getApiBaseURL(): string {
  if (typeof window !== 'undefined') {
    const configured = window.localStorage.getItem('outpipe_api_url');
    if (configured) return normalizeApiBaseURL(configured);
  }
  return normalizeApiBaseURL(
    env.VITE_OUTPIPE_API_BASE_URL ?? 'http://localhost:8080',
  );
}

export function setApiBaseURL(url: string): void {
  const normalized = normalizeApiBaseURL(url);
  if (typeof window !== 'undefined') {
    if (normalized) window.localStorage.setItem('outpipe_api_url', normalized);
    else window.localStorage.removeItem('outpipe_api_url');
  }
  axiosClient.defaults.baseURL = normalized || getApiBaseURL();
}

function normalizeError(error: unknown): ApiError {
  if (error instanceof ApiError) return error;
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as ApiErrorResponse | undefined;
    return new ApiError(
      error.response?.status ?? 0,
      data?.error ?? data?.message ?? error.message,
      data,
    );
  }
  return new ApiError(
    0,
    error instanceof Error ? error.message : 'Unexpected API error',
  );
}

export function createApiClient(): AxiosInstance {
  const client = axios.create({
    baseURL: getApiBaseURL(),
    withCredentials: true,
    headers: { Accept: 'application/json' },
  });
  client.interceptors.response.use(
    (response) => response,
    (error: unknown) => Promise.reject(normalizeError(error)),
  );
  return client;
}

export const axiosClient = createApiClient();

export const apiClient = {
  request<T>(config: AxiosRequestConfig) {
    return axiosClient.request<T>(config).then((response) => response.data);
  },
  get<T>(path: string, config?: AxiosRequestConfig) {
    return apiClient.request<T>({ ...config, method: 'GET', url: path });
  },
  post<T>(path: string, data?: unknown, config?: AxiosRequestConfig) {
    return apiClient.request<T>({ ...config, method: 'POST', url: path, data });
  },
  put<T>(path: string, data?: unknown, config?: AxiosRequestConfig) {
    return apiClient.request<T>({ ...config, method: 'PUT', url: path, data });
  },
  patch<T>(path: string, data?: unknown, config?: AxiosRequestConfig) {
    return apiClient.request<T>({
      ...config,
      method: 'PATCH',
      url: path,
      data,
    });
  },
  delete<T>(path: string, config?: AxiosRequestConfig) {
    return apiClient.request<T>({ ...config, method: 'DELETE', url: path });
  },
};

export function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const data = parseRequestBody(init?.body);
  return apiClient.request<T>({
    method: init?.method ?? 'GET',
    url: path,
    data,
    headers: init?.headers as AxiosRequestConfig['headers'],
    signal: init?.signal ?? undefined,
  });
}

function normalizeApiBaseURL(url: string): string {
  return url
    .trim()
    .replace(/\/+$/, '')
    .replace(/\/api$/, '');
}

function parseRequestBody(body: BodyInit | null | undefined): unknown {
  if (typeof body !== 'string') return body;
  try {
    return JSON.parse(body);
  } catch {
    return body;
  }
}
