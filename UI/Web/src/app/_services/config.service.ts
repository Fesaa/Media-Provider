import {Injectable} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {Config} from '../_models/config';
import {of, tap} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class ConfigService {

  baseUrl = environment.apiUrl + 'config/';
  syncId: number = -1;
  config: Config | undefined;

  constructor(private httpClient: HttpClient) {
  }


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
    return this.httpClient.get<string>(this.baseUrl + 'user/refresh-api-key');
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
