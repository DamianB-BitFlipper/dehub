import { type CSSProperties, useCallback, useEffect, useRef, useState } from 'react';

// Pixels per second the marquee text travels while hovered. Longer names take
// proportionally longer so the perceived speed stays constant.
const MARQUEE_SPEED_PX_PER_SEC = 40;
const MIN_DURATION_SEC = 2;

interface MarqueeState {
  containerRef: (node: HTMLElement | null) => void;
  trackRef: (node: HTMLElement | null) => void;
  overflowing: boolean;
  style: CSSProperties;
}

// Detects whether the inner track overflows its container and, when it does,
// exposes the CSS custom properties used by the hover marquee animation.
export function useOverflowMarquee(text: string): MarqueeState {
  const containerEl = useRef<HTMLElement | null>(null);
  const trackEl = useRef<HTMLElement | null>(null);
  const [overflow, setOverflow] = useState(0);

  const measure = useCallback(() => {
    const container = containerEl.current;
    const track = trackEl.current;
    if (container == null || track == null) {
      setOverflow(0);
      return;
    }
    const diff = track.scrollWidth - container.clientWidth;
    setOverflow(diff > 1 ? diff : 0);
  }, []);

  const containerRef = useCallback(
    (node: HTMLElement | null) => {
      containerEl.current = node;
      measure();
    },
    [measure]
  );

  const trackRef = useCallback(
    (node: HTMLElement | null) => {
      trackEl.current = node;
      measure();
    },
    [measure]
  );

  useEffect(() => {
    measure();
  }, [measure, text]);

  useEffect(() => {
    const container = containerEl.current;
    if (container == null || typeof ResizeObserver === 'undefined') {
      return undefined;
    }
    const observer = new ResizeObserver(() => measure());
    observer.observe(container);
    if (trackEl.current != null) {
      observer.observe(trackEl.current);
    }
    return () => observer.disconnect();
  }, [measure]);

  const durationSec = Math.max(
    MIN_DURATION_SEC,
    overflow / MARQUEE_SPEED_PX_PER_SEC
  );

  return {
    containerRef,
    trackRef,
    overflowing: overflow > 0,
    style: {
      '--marquee-shift': `-${overflow}px`,
      '--marquee-duration': `${durationSec.toFixed(2)}s`,
    } as CSSProperties,
  };
}
