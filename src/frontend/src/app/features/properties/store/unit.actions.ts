import { createActionGroup, props } from '@ngrx/store';
import { Unit, CreateUnitRequest, UpdateUnitRequest } from '../../../core/models/unit.model';

export const UnitActions = createActionGroup({
  source: 'Unit',
  events: {
    'Load Units': props<{ buildingId?: number; active?: boolean }>(),
    'Load Units Success': props<{ units: Unit[] }>(),
    'Load Units Failure': props<{ error: unknown }>(),

    'Load Unit': props<{ id: number }>(),
    'Load Unit Success': props<{ unit: Unit }>(),
    'Load Unit Failure': props<{ error: unknown }>(),

    'Create Unit': props<{ request: CreateUnitRequest }>(),
    'Create Unit Success': props<{ unit: Unit }>(),
    'Create Unit Failure': props<{ error: unknown }>(),

    'Update Unit': props<{ id: number; request: UpdateUnitRequest }>(),
    'Update Unit Success': props<{ unit: Unit }>(),
    'Update Unit Failure': props<{ error: unknown }>(),

    'Delete Unit': props<{ id: number }>(),
    'Delete Unit Success': props<{ id: number }>(),
    'Delete Unit Failure': props<{ error: unknown }>(),

    'Select Unit': props<{ id: number | null }>(),
  },
});
