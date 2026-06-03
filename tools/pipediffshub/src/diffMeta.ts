import { useEffect, useState } from 'react';

const DEFAULT_TITLE = 'PipeDiffsHub';

export interface DiffMeta {
  title: string;
  baseRefName: string;
  headRefName: string;
}

export function useDiffMeta(): DiffMeta {
  const [meta, setMeta] = useState<DiffMeta>({
    title: '',
    baseRefName: '',
    headRefName: '',
  });

  useEffect(() => {
    const controller = new AbortController();

    async function updateMeta() {
      try {
        const response = await fetch('/meta', {
          cache: 'no-store',
          signal: controller.signal,
        });
        if (!response.ok) return;

        const nextMeta = parseDiffMeta(await response.json());
        setMeta(nextMeta);
        const title = nextMeta.title.trim();
        document.title = title === '' ? DEFAULT_TITLE : title;
      } catch (error) {
        if (!controller.signal.aborted) {
          document.title = DEFAULT_TITLE;
        }
      }
    }

    updateMeta();
    return () => controller.abort();
  }, []);

  return meta;
}

function parseDiffMeta(value: unknown): DiffMeta {
  if (value == null || typeof value !== 'object') {
    return { title: '', baseRefName: '', headRefName: '' };
  }

  const record = value as Record<string, unknown>;
  return {
    title: typeof record.title === 'string' ? record.title : '',
    baseRefName:
      typeof record.baseRefName === 'string' ? record.baseRefName : '',
    headRefName:
      typeof record.headRefName === 'string' ? record.headRefName : '',
  };
}
