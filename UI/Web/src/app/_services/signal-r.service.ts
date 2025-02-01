import {Injectable} from '@angular/core';
import {HubConnection, HubConnectionBuilder} from "@microsoft/signalr";
import {environment} from "../../environments/environment";
import {User} from "../_models/user";

@Injectable({
  providedIn: 'root'
})
export class SignalRService {
  baseUrl = environment.apiUrl;
  private hubConnection!: HubConnection;

  constructor() {

  }

  startConnection(user: User) {
    this.hubConnection = new HubConnectionBuilder()
      .withUrl(this.baseUrl.substring(0, this.baseUrl.length - "api/".length) + "ws", {
        accessTokenFactory: () => user.token
      })
      .withAutomaticReconnect()
      .build()

    this.hubConnection
      .start()
      .catch((error) => {
        console.error('Error connecting to SignalR hub:', error);
      });
  }
}
