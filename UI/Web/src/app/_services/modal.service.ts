import {inject, Injectable, TemplateRef, Type} from '@angular/core';
import {NgbModal, NgbModalOptions, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {DefaultModalOptions} from '../_models/default-modal-options';
import {firstValueFrom, take} from 'rxjs';
import {ConfirmModalComponent} from "../shared/_component/confirm-modal/confirm-modal.component";


@Injectable({
  providedIn: 'root'
})
export class ModalService {

  private modal = inject(NgbModal);

  open<T>(content: Type<T>, options?: NgbModalOptions): [NgbModalRef, T]  {
    const modal = this.modal.open(content, options);
    return [modal, modal.componentInstance as T]
  }

  hasOpenModals() {
    return this.modal.hasOpenModals()
  }

  get activeInstances() {
    return this.modal.activeInstances
  }

  dismissAll(reason?: any) {
    this.modal.dismissAll(reason);
  }

  confirm(options: {
    question?: string;
    title?: string;
    bodyTemplate?: TemplateRef<unknown>;
    templateData?: unknown;
  }) {
    const [_, component] = this.open(ConfirmModalComponent, DefaultModalOptions);

    if (options.question) {
      component.question.set(options.question);
    }

    if (options.title) {
      component.title.set(options.title);
    }

    if (options.bodyTemplate) {
      component.bodyTemplate.set(options.bodyTemplate);
    }

    if (options.templateData) {
      component.templateData.set(options.templateData);
    }

    return firstValueFrom(component.result$);
  }


}
