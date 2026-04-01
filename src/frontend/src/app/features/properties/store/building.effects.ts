import { Injectable, inject } from '@angular/core'; // Wait, inject is from @angular/core
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, mergeMap, catchError } from 'rxjs/operators';
import { BuildingService } from '../../../core/services/building.service';
import { BuildingActions } from './building.actions';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Building, BuildingListResponse } from '../../../core/models';

@Injectable()
export class BuildingEffects {
  private actions$ = inject(Actions);
  private buildingService = inject(BuildingService);
  private snackBar = inject(MatSnackBar);

  loadBuildings$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.loadBuildings),
      mergeMap(({ propertyId, active }) => {
        if (propertyId) {
          return this.buildingService.getBuildingsByProperty(propertyId).pipe(
            map((response: BuildingListResponse) =>
              BuildingActions.loadBuildingsSuccess({ buildings: response.buildings })
            ),
            catchError((error) => of(BuildingActions.loadBuildingsFailure({ error })))
          );
        } else {
          return this.buildingService.getBuildings({ active }).pipe(
            map((response: BuildingListResponse) =>
              BuildingActions.loadBuildingsSuccess({ buildings: response.buildings })
            ),
            catchError((error) => of(BuildingActions.loadBuildingsFailure({ error })))
          );
        }
      })
    )
  );

  loadBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.loadBuilding),
      mergeMap(({ id }) =>
        this.buildingService.getBuilding(id).pipe(
          map((building: Building) => BuildingActions.loadBuildingSuccess({ building })),
          catchError((error) => of(BuildingActions.loadBuildingFailure({ error })))
        )
      )
    )
  );

  createBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.createBuilding),
      mergeMap(({ request }) =>
        this.buildingService.createBuilding(request).pipe(
          map((building: Building) => {
            this.snackBar.open('Building created successfully', 'Close', { duration: 3000 });
            return BuildingActions.createBuildingSuccess({ building });
          }),
          catchError((error) => {
            this.snackBar.open('Error creating building', 'Close', { duration: 3000 });
            return of(BuildingActions.createBuildingFailure({ error }));
          })
        )
      )
    )
  );

  updateBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.updateBuilding),
      mergeMap(({ id, request }) =>
        this.buildingService.updateBuilding(id, request).pipe(
          map((building: Building) => {
            this.snackBar.open('Building updated successfully', 'Close', { duration: 3000 });
            return BuildingActions.updateBuildingSuccess({ building });
          }),
          catchError((error) => {
            this.snackBar.open('Error updating building', 'Close', { duration: 3000 });
            return of(BuildingActions.updateBuildingFailure({ error }));
          })
        )
      )
    )
  );

  deleteBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.deleteBuilding),
      mergeMap(({ id }) =>
        this.buildingService.deleteBuilding(id).pipe(
          map(() => {
            this.snackBar.open('Building deleted successfully', 'Close', { duration: 3000 });
            return BuildingActions.deleteBuildingSuccess({ id });
          }),
          catchError((error) => {
            this.snackBar.open('Error deleting building', 'Close', { duration: 3000 });
            return of(BuildingActions.deleteBuildingFailure({ error }));
          })
        )
      )
    )
  );
}
