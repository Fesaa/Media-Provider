import {ChangeDetectionStrategy, Component, inject} from '@angular/core';
import {ContentService} from "../../../_services/content.service";
import {ToastService} from "../../../_services/toast.service";
import {NgbActiveModal} from "@ng-bootstrap/ng-bootstrap";

@Component({
  selector: 'app-manual-content-add-modal',
  imports: [],
  templateUrl: './manual-content-add-modal.component.html',
  styleUrl: './manual-content-add-modal.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ManualContentAddModalComponent {

  private readonly contentService = inject(ContentService);
  private readonly toastService = inject(ToastService);
  private readonly modal = inject(NgbActiveModal);

  close() {}

  submit() {}

}
