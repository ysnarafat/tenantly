import { createActionGroup, props } from '@ngrx/store';
import {
  Building,
  CreateBuildingRequest,
  UpdateBuildingRequest,
} from '../../../core/models/building.model';

export const BuildingActions = createActionGroup({
  source: 'Building',
  events: {
    'Load Buildings': props<{ propertyId?: number; active?: boolean }>(),
    'Load Buildings Success': props<{ buildings: Building[] }>(),
    'Load Buildings Failure': props<{ error: unknown }>(),

    'Load Building': props<{ id: number }>(),
    'Load Building Success': props<{ building: Building }>(),
    'Load Building Failure': props<{ error: unknown }>(),

    'Create Building': props<{ request: CreateBuildingRequest }>(),
    'Create Building Success': props<{ building: Building }>(),
    'Create Building Failure': props<{ error: unknown }>(),

    'Update Building': props<{ id: number; request: UpdateBuildingRequest }>(),
    'Update Building Success': props<{ building: Building }>(),
    'Update Building Failure': props<{ error: unknown }>(),

    'Delete Building': props<{ id: number }>(),
    'Delete Building Success': props<{ id: number }>(),
    'Delete Building Failure': props<{ error: unknown }>(),

    'Select Building': props<{ id: number | null }>(),
  },
});
