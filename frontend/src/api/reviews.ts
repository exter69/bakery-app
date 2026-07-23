import { apiFetch } from './client';
import type { ListResponse } from '../types/bakery';

export interface ReviewResponse {
  id: string;
  rating: number;
  text: string | null;
  authorName: string;
  createdAt: string;
}

export function fetchReviews(bakeryId: string, page?: number): Promise<ListResponse<ReviewResponse>> {
  const params = new URLSearchParams();
  if (page) params.set('page', String(page));
  return apiFetch<ListResponse<ReviewResponse>>(`/bakeries/${bakeryId}/reviews?${params}`);
}

export function createReview(bakeryId: string, data: { rating: number; text?: string }): Promise<ReviewResponse> {
  return apiFetch<ReviewResponse>(`/bakeries/${bakeryId}/reviews`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function reportReview(reviewId: string, reason: string): Promise<void> {
  return apiFetch<void>(`/reviews/${reviewId}/report`, {
    method: 'POST',
    body: JSON.stringify({ reason }),
  });
}
