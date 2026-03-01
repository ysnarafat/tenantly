import { createReducer, on } from '@ngrx/store';
import { EntityState, EntityAdapter, createEntityAdapter } from '@ngrx/entity';
import { Property } from '../../../core/models';
import { PropertyActions } from './property.actions';

export interface PropertyState extends EntityState<Property> {
  selectedId: number | null;
  loading: boolean;
  error: string | null;
}

export const adapter: EntityAdapter<Property> = createEntityAdapter<Property>();

export const initialState: PropertyState = adapter.getInitialState({
  selectedId: null,
  loading: false,
  error: null,
});

export const propertyReducer = createReducer(
  initialState,
  // Load Properties
  on(PropertyActions.loadProperties, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(PropertyActions.loadPropertiesSuccess, (state, { properties }) =>
    adapter.setAll(properties, { ...state, loading: false })
  ),
  on(PropertyActions.loadPropertiesFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Load Property
  on(PropertyActions.loadProperty, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(PropertyActions.loadPropertySuccess, (state, { property }) =>
    adapter.upsertOne(property, { ...state, loading: false, selectedId: property.id })
  ),
  on(PropertyActions.loadPropertyFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Create Property
  on(PropertyActions.createProperty, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(PropertyActions.createPropertySuccess, (state, { property }) =>
    adapter.addOne(property, { ...state, loading: false })
  ),
  on(PropertyActions.createPropertyFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Update Property
  on(PropertyActions.updateProperty, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(PropertyActions.updatePropertySuccess, (state, { property }) =>
    adapter.updateOne({ id: property.id, changes: property }, { ...state, loading: false })
  ),
  on(PropertyActions.updatePropertyFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Delete Property
  on(PropertyActions.deleteProperty, (state) => ({
    ...state,
    loading: true,
    error: null,
  })),
  on(PropertyActions.deletePropertySuccess, (state, { id }) =>
    adapter.removeOne(id, { ...state, loading: false })
  ),
  on(PropertyActions.deletePropertyFailure, (state, { error }) => ({
    ...state,
    loading: false,
    error,
  })),

  // Selection
  on(PropertyActions.selectProperty, (state, { id }) => ({
    ...state,
    selectedId: id,
  })),
  on(PropertyActions.clearSelectedProperty, (state) => ({
    ...state,
    selectedId: null,
  }))
);

export const { selectIds, selectEntities, selectAll, selectTotal } = adapter.getSelectors();
