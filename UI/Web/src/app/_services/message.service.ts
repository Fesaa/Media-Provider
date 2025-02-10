import {Injectable} from '@angular/core';
import {MessageService as PrimeNgMessageService, ToastMessageOptions} from "primeng/api";

@Injectable({
  providedIn: 'root'
})
export class MessageService {

  constructor(private msgService: PrimeNgMessageService) {
  }

  info(title: string, message?: string, opts?: ToastMessageOptions) {
    this.msgService.add({
      summary: title,
      detail: message,
      severity: 'info',
      ...opts
    })
  }

  success(title: string, message?: string, opts?: ToastMessageOptions) {
    this.msgService.add({
      summary: title,
      detail: message,
      severity: 'success',
      ...opts
    })
  }

  warning(title: string, message?: string, opts?: ToastMessageOptions) {
    this.msgService.add({
      summary: title,
      detail: message,
      severity: 'warn',
      ...opts
    })
  }

  error(title: string, message?: string, opts?: ToastMessageOptions) {
    console.debug(`An error occurred${title}:\n ${message}`);
    this.msgService.add({
      summary: title,
      detail: message,
      severity: 'error',
      ...opts
    })
  }

}
