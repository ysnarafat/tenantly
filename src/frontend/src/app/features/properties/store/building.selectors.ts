import { createFeatureSelector, createSelector } from '@ngrx/store';
import { BuildingState, buildingAdapter } from './building.reducer';

export const selectBuildingState = createFeatureSelector<BuildingState>('buildings');

const { selectAll, selectEntities, selectIds, selectTotal } = buildingAdapter.getSelectors();

export const selectAllBuildings = createSelector(selectBuildingState, selectAll);

export const selectBuildingEntities = createSelector(selectBuildingState, selectEntities);

export const selectBuildingLoading = createSelector(selectBuildingState, (state) => state.loading);

export const selectBuildingError = createSelector(selectBuildingState, (state) => state.error);

export const selectSelectedBuildingId = createSelector(
  selectBuildingState,
  (state) => state.selectedId
);

export const selectSelectedBuilding = createSelector(
  selectBuildingEntities,
  selectSelectedBuildingId,
  (entities, selectedId) => (selectedId ? entities[selectedId] : null)
);
