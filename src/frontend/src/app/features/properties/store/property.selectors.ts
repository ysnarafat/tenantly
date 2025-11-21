import { createFeatureSelector, createSelector } from '@ngrx/store';
import { PropertyState, selectAll, selectEntities } from './property.reducer';

export const selectPropertyState = createFeatureSelector<PropertyState>('properties');

export const selectAllProperties = createSelector(
    selectPropertyState,
    selectAll
);

export const selectPropertyEntities = createSelector(
    selectPropertyState,
    selectEntities
);

export const selectPropertyLoading = createSelector(
    selectPropertyState,
    (state) => state.loading
);

export const selectPropertyError = createSelector(
    selectPropertyState,
    (state) => state.error
);

export const selectSelectedPropertyId = createSelector(
    selectPropertyState,
    (state) => state.selectedId
);

export const selectSelectedProperty = createSelector(
    selectPropertyEntities,
    selectSelectedPropertyId,
    (entities, selectedId) => selectedId ? entities[selectedId] : null
);
