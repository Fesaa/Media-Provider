import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {Config, MovePageRequest} from '../_models/config';
import {Page} from "../_models/page";

@Injectable({
  providedIn: 'root'
})
export class ConfigService {

  baseUrl = environment.apiUrl + 'config/';

  constructor(private httpClient: HttpClient) { }


  getConfig() {
    return this.httpClient.get<Config>(this.baseUrl);
  }

  removePage(pageId: number, syncId: number) {
    return this.httpClient.delete<number>(this.baseUrl + 'pages/' + pageId + '?sync_id' + syncId);
  }

  addPage(page: Page, syncId: number) {
    return this.httpClient.post<number>(this.baseUrl + 'pages?sync_id=' + syncId, page);
  }

  updatePage(page: Page, pageIndex: number, syncId: number) {
    return this.httpClient.put<number>(this.baseUrl + 'pages/' + pageIndex + '?sync_id=' + syncId, page);
  }

  movePage(oldIndex: number, newIndex: number, syncId: number) {
    const req: MovePageRequest = {
      oldIndex: oldIndex,
      newIndex: newIndex
    };
    return this.httpClient.post<number>(this.baseUrl + 'pages/move?sync_id=' + syncId, req);
  }

}
