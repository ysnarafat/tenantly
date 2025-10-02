import { createAction, props } from '@ngrx/store';

// Generic loading actions
export const createLoadingActions = (feature: string) => ({
  setLoading: createAction(
    `[${feature}] Set Loading`,
    props<{ loading: boolean }>()
  ),
  clearError: createAction(`[${feature}] Clear Error`),
});

// Generic CRUD actions factory
export const createCrudActions = <T>(feature: string) => ({
  // Load actions
  load: createAction(`[${feature}] Load`),
  loadSuccess: createAction(
    `[${feature}] Load Success`,
    props<{ items: T[] }>()
  ),
  loadFailure: createAction(
    `[${feature}] Load Failure`,
    props<{ error: any }>()
  ),

  // Create actions
  create: createAction(
    `[${feature}] Create`,
    props<{ item: Partial<T> }>()
  ),
  createSuccess: createAction(
    `[${feature}] Create Success`,
    props<{ item: T }>()
  ),
  createFailure: createAction(
    `[${feature}] Create Failure`,
    props<{ error: any }>()
  ),

  // Update actions
  update: createAction(
    `[${feature}] Update`,
    props<{ id: string | number; changes: Partial<T> }>()
  ),
  updateSuccess: createAction(
    `[${feature}] Update Success`,
    props<{ item: T }>()
  ),
  updateFailure: createAction(
    `[${feature}] Update Failure`,
    props<{ error: any }>()
  ),

  // Delete actions
  delete: createAction(
    `[${feature}] Delete`,
    props<{ id: string | number }>()
  ),
  deleteSuccess: createAction(
    `[${feature}] Delete Success`,
    props<{ id: string | number }>()
  ),
  deleteFailure: createAction(
    `[${feature}] Delete Failure`,
    props<{ error: any }>()
  ),

  // Select actions
  select: createAction(
    `[${feature}] Select`,
    props<{ id: string | number | null }>()
  ),
});