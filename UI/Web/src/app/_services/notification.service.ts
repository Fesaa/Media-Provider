import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import {environment} from "../../environments/environment";
import {Notification} from "../_models/notifications";

@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  private baseUrl = environment.apiUrl + "notifications";

  constructor(private http: HttpClient) {}

  all(after?: Date) {
    let params = new HttpParams();
    if (after) {
      const formattedDate = after.toISOString();
      params = params.set('after', formattedDate);
    }
    return this.http.get<Notification[]>(`${this.baseUrl}/all`, { params });
  }

  amount() {
    return this.http.get<number>(`${this.baseUrl}/amount`);
  }

  markAsRead(id: number) {
    return this.http.post<any>(`${this.baseUrl}/${id}/read`, {});
  }

  markAsUnread(id: number) {
    return this.http.post<any>(`${this.baseUrl}/${id}/unread`, {});
  }

  deleteNotification(id: number) {
    return this.http.delete<any>(`${this.baseUrl}/${id}`);
  }
}
