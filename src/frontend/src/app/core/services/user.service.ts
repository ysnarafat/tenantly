import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { User } from './auth.service';

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  role: string;
  organization_id?: number;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  first_name?: string;
  last_name?: string;
  role?: string;
  status?: string;
  active?: boolean;
}

interface UsersResponse {
  users: User[];
  total?: number;
}

@Injectable({
  providedIn: 'root',
})
export class UserService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/users`;

  getUsers(activeOnly = true): Observable<UsersResponse> {
    let params = new HttpParams();
    if (!activeOnly) {
      params = params.set('active', 'false');
    }
    return this.http.get<UsersResponse>(this.apiUrl, { params });
  }

  getUser(id: number): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/${id}`);
  }

  createUser(req: CreateUserRequest): Observable<User> {
    return this.http.post<User>(this.apiUrl, req);
  }

  updateUser(id: number, req: UpdateUserRequest): Observable<{ message: string }> {
    return this.http.put<{ message: string }>(`${this.apiUrl}/${id}`, req);
  }

  deleteUser(id: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/${id}`);
  }

  adminResetPassword(id: number, newPassword: string): Observable<{ message: string }> {
    return this.http.put<{ message: string }>(`${this.apiUrl}/${id}/reset-password`, {
      new_password: newPassword,
    });
  }
}
