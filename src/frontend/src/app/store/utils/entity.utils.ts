import { EntityAdapter, createEntityAdapter } from '@ngrx/entity';

// Generic entity adapter factory
export function createGenericEntityAdapter<T extends { id: number | string }>(): EntityAdapter<T> {
  return createEntityAdapter<T>({
    selectId: (entity: T) => entity.id,
    sortComparer: false,
  });
}

// Entity state interface
export interface EntityState<T> {
  ids: (string | number)[];
  entities: Record<string | number, T>;
  loading: boolean;
  error: any;
  selectedId: string | number | null;
}

// Initial entity state factory
export function createInitialEntityState<T>(): EntityState<T> {
  return {
    ids: [],
    entities: {},
    loading: false,
    error: null,
    selectedId: null,
  };
}

// Loading state helpers
export interface LoadingState {
  loading: boolean;
  error: any;
}

export const createLoadingState = (): LoadingState => ({
  loading: false,
  error: null,
});

export const setLoading = <T extends LoadingState>(state: T): T => ({
  ...state,
  loading: true,
  error: null,
});

export const setLoaded = <T extends LoadingState>(state: T): T => ({
  ...state,
  loading: false,
  error: null,
});

export const setError = <T extends LoadingState>(state: T, error: any): T => ({
  ...state,
  loading: false,
  error,
});
