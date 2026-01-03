import { createFeatureSelector, createSelector } from '@ngrx/store';
import { UnitState, unitAdapter } from './unit.reducer';

export const selectUnitState = createFeatureSelector<UnitState>('units');

const { selectAll, selectEntities, selectIds, selectTotal } = unitAdapter.getSelectors();

export const selectAllUnits = createSelector(selectUnitState, selectAll);

export const selectUnitEntities = createSelector(selectUnitState, selectEntities);

export const selectUnitLoading = createSelector(selectUnitState, (state) => state.loading);

export const selectUnitError = createSelector(selectUnitState, (state) => state.error);

export const selectSelectedUnitId = createSelector(selectUnitState, (state) => state.selectedId);

export const selectSelectedUnit = createSelector(
  selectUnitEntities,
  selectSelectedUnitId,
  (entities, selectedId) => (selectedId ? entities[selectedId] : null)
);
