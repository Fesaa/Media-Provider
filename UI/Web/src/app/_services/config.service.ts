import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {Config, MovePageRequest} from '../_models/config';
import {Page} from "../_models/page";
import {of, Subject, take, tap} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class ConfigService {

  baseUrl = environment.apiUrl + 'config/';
  syncId: number = -1;
  config: Config | undefined;

  constructor(private httpClient: HttpClient) { }


  getConfig() {
    if (this.config && this.config.sync_id == this.syncId) {
      return of(this.config);
    }

    return this.httpClient.get<Config>(this.baseUrl).pipe(tap(config => {
      this.config = config;
      this.syncId = config.sync_id;
    }));
  }

  refreshApiKey() {
    if (this.syncId == -1) {
      throw new Error('Sync ID is not set');
    }

    const subject = new Subject<string>();
    this.httpClient.get<{apiKey: string, sync_id: number}>(this.baseUrl + 'refresh-api-key' + '?sync_id=' + this.syncId)
      .subscribe(model => {
        this.syncId = model.sync_id;
        subject.next(model.apiKey);
      })
    return subject.asObservable();
  }

  removePage(pageId: number) {
    if (this.syncId == -1) {
      throw new Error('Sync ID is not set');
    }

    return this.httpClient
      .delete<number>(this.baseUrl + 'pages/' + pageId + '?sync_id=' + this.syncId)
      .pipe(this.updateSyncId());
  }

  addPage(page: Page) {
    if (this.syncId == -1) {
      throw new Error('Sync ID is not set');
    }
    return this.httpClient
      .post<number>(this.baseUrl + 'pages?sync_id=' + this.syncId, page)
      .pipe(this.updateSyncId());
  }

  updatePage(page: Page, pageIndex: number) {
    if (this.syncId == -1) {
      throw new Error('Sync ID is not set');
    }
    return this.httpClient
      .put<number>(this.baseUrl + 'pages/' + pageIndex + '?sync_id=' + this.syncId, page)
      .pipe(this.updateSyncId());
  }

  movePage(oldIndex: number, newIndex: number) {
    if (this.syncId == -1) {
      throw new Error('Sync ID is not set');
    }
    const req: MovePageRequest = {
      oldIndex: oldIndex,
      newIndex: newIndex
    };
    return this.httpClient
      .post<number>(this.baseUrl + 'pages/move?sync_id=' + this.syncId, req)
      .pipe(this.updateSyncId());
  }

  updateConfig(config: Config) {
    if (this.syncId == -1) {
      throw new Error('Sync ID is not set');
    }
    return this.httpClient
      .post<number>(this.baseUrl + 'update?sync_id=' + this.syncId, config)
      .pipe(this.updateSyncId());
  }

  private updateSyncId() {
    return tap((syncId: number) => {
      this.syncId = syncId;
    })
  }

}
