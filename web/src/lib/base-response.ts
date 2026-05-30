export interface ApiError {
  code: number;
  message: string;
}

export interface PaginatedMeta {
  limit: number;
  page: number;
  total: number;
  totalPages: number;
}

export interface BaseResponse<T> {
  data?: T | null;
  error?: ApiError | null;
  meta?: PaginatedMeta;
}