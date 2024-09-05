import {
  Injectable,
  ComponentRef,
  ApplicationRef,
  ViewContainerRef
} from '@angular/core';
import {DialogComponent} from "../shared/_component/dialog/dialog.component";

@Injectable({
  providedIn: 'root'
})
export class DialogService {

  public viewContainerRef: ViewContainerRef | undefined;

  constructor(
    private appRef: ApplicationRef,
  ) {}

  openDialog(text: string): Promise<boolean> {
    const component = this.viewContainerRef!.createComponent(DialogComponent)
    component.instance.text = text;

    return new Promise<boolean>((resolve, reject) => {
      component.instance.getResult().subscribe(result => {
        this.closeDialog(component);
        resolve(result);
      });
    });
  }

  closeDialog(componentRef: ComponentRef<DialogComponent>) {
    this.appRef.detachView(componentRef.hostView);
    componentRef.destroy();
  }
}
