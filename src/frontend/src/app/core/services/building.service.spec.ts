import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { BuildingService } from './building.service';
import { BuildingWithStats } from '../models';
import { environment } from '../../../environments/environment';

describe('BuildingService', () => {
  let service: BuildingService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/buildings`;

  const mockBuildingWithStats: BuildingWithStats = {
    id: 1,
    property_id: 1,
    building_name: 'Block A',
    building_code: 'BLK-A',
    building_type: 'Residential',
    has_elevator: true,
    active_status: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    property_name: 'Sunrise Apartments',
    unit_count: 10,
    occupied_units: 8,
    total_revenue: 200000,
    occupancy_rate: 80,
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [BuildingService],
    });
    service = TestBed.inject(BuildingService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('getBuildingWithStats', () => {
    it('GETs the building by id with the include_stats param', () => {
      service.getBuildingWithStats(1).subscribe();

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/1`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('include_stats')).toBe('true');
      req.flush({ building: mockBuildingWithStats });
    });

    it('unwraps the {building: ...} envelope the backend returns', () => {
      let result: BuildingWithStats | undefined;
      service.getBuildingWithStats(1).subscribe((res) => (result = res));

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/1`);
      req.flush({ building: mockBuildingWithStats });

      expect(result).toEqual(mockBuildingWithStats);
    });

    it('propagates a 404 when the building does not exist', () => {
      service.getBuildingWithStats(999).subscribe({
        next: () => fail('expected error'),
        error: (err) => expect(err.status).toBe(404),
      });

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });
  });
});
