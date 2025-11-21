import { createReducer, on } from '@ngrx/store';
import { EntityState, EntityAdapter, createEntityAdapter } from '@ngrx/entity';
import { Unit } from '../../../core/models/unit.model';
import { UnitActions } from './unit.actions';

export interface UnitState extends EntityState<Unit> {
    selectedId: number | null;
    loading: boolean;
    error: any;
}

export const unitAdapter: EntityAdapter<Unit> = createEntityAdapter<Unit>();

export const initialUnitState: UnitState = unitAdapter.getInitialState({
    selectedId: null,
    loading: false,
    error: null
});

export const unitReducer = createReducer(
    initialUnitState,

    // Load Units
    on(UnitActions.loadUnits, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(UnitActions.loadUnitsSuccess, (state, { units }) =>
        unitAdapter.setAll(units, { ...state, loading: false })
    ),
    on(UnitActions.loadUnitsFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Load Unit
    on(UnitActions.loadUnit, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(UnitActions.loadUnitSuccess, (state, { unit }) =>
        unitAdapter.upsertOne(unit, { ...state, loading: false, selectedId: unit.id })
    ),
    on(UnitActions.loadUnitFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Create Unit
    on(UnitActions.createUnit, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(UnitActions.createUnitSuccess, (state, { unit }) =>
        unitAdapter.addOne(unit, { ...state, loading: false })
    ),
    on(UnitActions.createUnitFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Update Unit
    on(UnitActions.updateUnit, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(UnitActions.updateUnitSuccess, (state, { unit }) =>
        unitAdapter.updateOne({ id: unit.id, changes: unit }, { ...state, loading: false })
    ),
    on(UnitActions.updateUnitFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Delete Unit
    on(UnitActions.deleteUnit, (state) => ({
        ...state,
        loading: true,
        error: null
    })),
    on(UnitActions.deleteUnitSuccess, (state, { id }) =>
        unitAdapter.removeOne(id, { ...state, loading: false })
    ),
    on(UnitActions.deleteUnitFailure, (state, { error }) => ({
        ...state,
        loading: false,
        error
    })),

    // Select Unit
    on(UnitActions.selectUnit, (state, { id }) => ({
        ...state,
        selectedId: id
    }))
);
