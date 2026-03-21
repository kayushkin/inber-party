import { useEffect, useRef, useState } from 'react';

export const useInViewAnimation = (threshold = 0.1) => {
  const [inView, setInView] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true);
          // Disconnect observer entirely after first intersection
          observer.disconnect();
        }
      },
      { threshold }
    );

    const currentRef = ref.current;
    if (currentRef) {
      observer.observe(currentRef);
    }

    return () => {
      // Clean up with captured ref to avoid stale reference
      observer.disconnect();
    };
  }, [threshold]);

  return { ref, inView };
};