import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {CreateDirRequest, DirEntry, ListDirRequest} from "../_models/io";
import {environment} from "../../environments/environment";

@Injectable({
  providedIn: 'root'
})
export class IoService {

  baseUrl = environment.apiUrl

  constructor(private httpClient: HttpClient) { }

  ls(req: ListDirRequest) {
    return this.httpClient.post<DirEntry[]>(this.baseUrl + 'io/ls', req);
  }

  create(req: CreateDirRequest) {
    return this.httpClient.post(this.baseUrl + 'io/create', req);
  }

}
