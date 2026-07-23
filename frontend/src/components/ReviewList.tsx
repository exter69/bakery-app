import { useCallback, useEffect, useState } from 'react';
import { fetchReviews, type ReviewResponse } from '../api/reviews';
import { useI18n } from '../i18n';
import StarRating from './StarRating';
import './ReviewList.css';

interface ReviewListProps {
  bakeryId: string;
}

function getRelativeTime(dateStr: string, locale: string): string {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diffMs = now - date;
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHr = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHr / 24);
  const diffWeek = Math.floor(diffDay / 7);
  const diffMonth = Math.floor(diffDay / 30);
  const diffYear = Math.floor(diffDay / 365);

  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });

  if (diffYear > 0) return rtf.format(-diffYear, 'year');
  if (diffMonth > 0) return rtf.format(-diffMonth, 'month');
  if (diffWeek > 0) return rtf.format(-diffWeek, 'week');
  if (diffDay > 0) return rtf.format(-diffDay, 'day');
  if (diffHr > 0) return rtf.format(-diffHr, 'hour');
  if (diffMin > 0) return rtf.format(-diffMin, 'minute');
  return rtf.format(-diffSec, 'second');
}

export default function ReviewList({ bakeryId }: ReviewListProps) {
  const { t, locale } = useI18n();
  const [reviews, setReviews] = useState<ReviewResponse[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);

  const loadReviews = useCallback(async (pageNum: number, append: boolean) => {
    try {
      const res = await fetchReviews(bakeryId, pageNum);
      setReviews((prev) => (append ? [...prev, ...res.items] : res.items));
      setTotal(res.total);
      setPage(pageNum);
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [bakeryId]);

  useEffect(() => {
    setLoading(true);
    loadReviews(1, false);
  }, [loadReviews]);

  const handleLoadMore = () => {
    setLoadingMore(true);
    loadReviews(page + 1, true);
  };

  const hasMore = reviews.length < total;

  if (loading) {
    return null;
  }

  return (
    <section className="review-list" aria-label={t('reviews.title')}>
      <h2 className="review-list__title">{t('reviews.title')}</h2>

      {reviews.length === 0 ? (
        <p className="review-list__empty">{t('reviews.empty')}</p>
      ) : (
        <>
          <div className="review-list__items">
            {reviews.map((review) => (
              <article key={review.id} className="review-list__item">
                <div className="review-list__item-header">
                  <StarRating rating={review.rating} size="sm" />
                  <span className="review-list__author">{review.authorName}</span>
                  <span className="review-list__time">
                    {getRelativeTime(review.createdAt, locale)}
                  </span>
                </div>
                {review.text && (
                  <p className="review-list__text">{review.text}</p>
                )}
              </article>
            ))}
          </div>

          {hasMore && (
            <button
              type="button"
              className="review-list__load-more"
              onClick={handleLoadMore}
              disabled={loadingMore}
            >
              {loadingMore ? '...' : t('history.next')}
            </button>
          )}
        </>
      )}
    </section>
  );
}
