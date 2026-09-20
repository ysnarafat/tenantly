import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { PropertyService } from './property.service';
import { PropertyWithStats } from '../models';
import { environment } from '../../../environments/environment';

describe('PropertyService', () => {
  let service: PropertyService;
  let httpMock: HttpTestingController;

  const apiUrl = `${environment.apiUrl}/properties`;

  const mockPropertyWithStats: PropertyWithStats = {
    id: 1,
    property_name: 'Sunrise Apartments',
    property_code: 'PROP-001',
    address: '12 Gulshan Ave',
    property_type: 'Residential',
    total_buildings: 2,
    active: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    building_count: 2,
    unit_count: 20,
    occupied_units: 15,
    total_revenue: 450000,
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [PropertyService],
    });
    service = TestBed.inject(PropertyService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('getPropertyWithStats', () => {
    it('GETs the property by id with include_stats and include_buildings params', () => {
      service.getPropertyWithStats(1).subscribe();

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/1`);
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('include_stats')).toBe('true');
      expect(req.request.params.get('include_buildings')).toBe('true');
      req.flush({ property: mockPropertyWithStats });
    });

    it('unwraps the {property: ...} envelope the backend returns', () => {
      let result: PropertyWithStats | undefined;
      service.getPropertyWithStats(1).subscribe((res) => (result = res));

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/1`);
      req.flush({ property: mockPropertyWithStats, building_summary: [{ id: 1 }] });

      expect(result).toEqual(mockPropertyWithStats);
    });

    it('propagates a 404 when the property does not exist', () => {
      service.getPropertyWithStats(999).subscribe({
        next: () => fail('expected error'),
        error: (err) => expect(err.status).toBe(404),
      });

      const req = httpMock.expectOne((r) => r.url === `${apiUrl}/999`);
      req.flush('Not found', { status: 404, statusText: 'Not Found' });
    });
  });
});
