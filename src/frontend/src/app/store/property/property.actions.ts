import { createAction, props } from '@ngrx/store';
import { Property, PropertyWithStats, CreatePropertyRequest, UpdatePropertyRequest } from '../../core/models';

// Load Properties
export const loadProperties = createAction(
  '[Property] Load Properties',
  props<{ active?: boolean }>()
);

export const loadPropertiesSuccess = createAction(
  '[Property] Load Properties Success',
  props<{ properties: Property[] }>()
);

export const loadPropertiesFailure = createAction(
  '[Property] Load Properties Failure',
  props<{ error: any }>()
);

// Load Single Property
export const loadProperty = createAction(
  '[Property] Load Property',
  props<{ id: number }>()
);

export const loadPropertySuccess = createAction(
  '[Property] Load Property Success',
  props<{ property: Property }>()
);

export const loadPropertyFailure = createAction(
  '[Property] Load Property Failure',
  props<{ error: any }>()
);

// Load Property with Stats
export const loadPropertyWithStats = createAction(
  '[Property] Load Property With Stats',
  props<{ id: number }>()
);

export const loadPropertyWithStatsSuccess = createAction(
  '[Property] Load Property With Stats Success',
  props<{ property: PropertyWithStats }>()
);

export const loadPropertyWithStatsFailure = createAction(
  '[Property] Load Property With Stats Failure',
  props<{ error: any }>()
);

// Create Property
export const createProperty = createAction(
  '[Property] Create Property',
  props<{ property: CreatePropertyRequest }>()
);

export const createPropertySuccess = createAction(
  '[Property] Create Property Success',
  props<{ property: Property }>()
);

export const createPropertyFailure = createAction(
  '[Property] Create Property Failure',
  props<{ error: any }>()
);

// Update Property
export const updateProperty = createAction(
  '[Property] Update Property',
  props<{ id: number; property: UpdatePropertyRequest }>()
);

export const updatePropertySuccess = createAction(
  '[Property] Update Property Success',
  props<{ property: Property }>()
);

export const updatePropertyFailure = createAction(
  '[Property] Update Property Failure',
  props<{ error: any }>()
);

// Delete Property
export const deleteProperty = createAction(
  '[Property] Delete Property',
  props<{ id: number }>()
);

export const deletePropertySuccess = createAction(
  '[Property] Delete Property Success',
  props<{ id: number }>()
);

export const deletePropertyFailure = createAction(
  '[Property] Delete Property Failure',
  props<{ error: any }>()
);

// Select Property
export const selectProperty = createAction(
  '[Property] Select Property',
  props<{ id: number | null }>()
);

// Clear Property State
export const clearPropertyState = createAction(
  '[Property] Clear Property State'
);
