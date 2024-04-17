import React, { ReactNode } from "react";
import SuccesNotification from "./succes";
import ErrorNotification from "./error";

class NotificationHandler extends React.Component {
  static notifications: React.Component[] = [];
  static fadeInTime: number = 2000;
  static instance: NotificationHandler | null = null;

  constructor(props: any) {
    super(props);
    NotificationHandler.instance = this;
  }

  private static forceUpdate() {
    if (NotificationHandler.instance) {
      NotificationHandler.instance.forceUpdate();
    }
  }

  public static addSuccesNotificationByTitle(title: string) {
    NotificationHandler.addNotification(
      new SuccesNotification({
        title: title,
        description: null,
      }),
    );
  }

  public static addErrorNotificationByTitle(title: string) {
    NotificationHandler.addNotification(
      new ErrorNotification({
        title: title,
        description: null,
      }),
    );
  }

  public static addNotification(notification: React.Component): void {
    NotificationHandler.notifications.push(notification);
    NotificationHandler.forceUpdate();

    setTimeout(() => {
      NotificationHandler.notifications.shift();
      NotificationHandler.forceUpdate();
    }, NotificationHandler.fadeInTime);
  }

  render(): React.ReactNode {
    return (
      <div className="fixed top-4 md:top-10 right-4 flex flex-row space-y-4">
        {NotificationHandler.notifications.map((comp) => comp.render())}
      </div>
    );
  }
}

export default NotificationHandler;
