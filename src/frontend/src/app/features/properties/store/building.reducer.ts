import { createReducer, on } from '@ngrx/store';
import { EntityState, EntityAdapter, createEntityAdapter } from '@ngrx/entity';
import { Building } from '../../../core/models/building.model';
import { BuildingActions } from './building.actions';

export interface BuildingState extends EntityState<Building> {
    selectedId: number | null;
    loading: boolean;
    error: any;
}

export const buildingAdapter: EntityAdapter<Building> = createEntityAdapter<Building>();

export const initialBuildingState: BuildingState = buildingAdapter.getInitialState({
    selectedId: null,
    loading: false,
    error: null
});

export const buildingReducer = createReducer(
    initialBuildingState,

    // Load Buildings
    on(BuildingActions.loadBuildings, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(BuildingActions.loadBuildingsSuccess, (state, { buildings }) =>
        buildingAdapter.setAll(buildings, { ...state, loading: false })
    ),
    on(BuildingActions.loadBuildingsFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Load Building
    on(BuildingActions.loadBuilding, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(BuildingActions.loadBuildingSuccess, (state, { building }) =>
        buildingAdapter.upsertOne(building, { ...state, loading: false, selectedId: building.id })
    ),
    on(BuildingActions.loadBuildingFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Create Building
    on(BuildingActions.createBuilding, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(BuildingActions.createBuildingSuccess, (state, { building }) =>
        buildingAdapter.addOne(building, { ...state, loading: false })
    ),
    on(BuildingActions.createBuildingFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Update Building
    on(BuildingActions.updateBuilding, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(BuildingActions.updateBuildingSuccess, (state, { building }) =>
        buildingAdapter.updateOne({ id: building.id, changes: building }, { ...state, loading: false })
    ),
    on(BuildingActions.updateBuildingFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Delete Building
    on(BuildingActions.deleteBuilding, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(BuildingActions.deleteBuildingSuccess, (state, { id }) =>
        buildingAdapter.removeOne(id, { ...state, loading: false })
    ),
    on(BuildingActions.deleteBuildingFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Select Building
    on(BuildingActions.selectBuilding, (state, { id }) => ({
        ...state,
        selectedId: id
    }))
);
