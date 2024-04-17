import React from "react";

class ErrorNotification extends React.Component {
  title: string;
  description: string | null;

  constructor(props: { title: string; description: string | null }) {
    super(props);
    this.title = props.title;
    this.description = props.description;
  }

  render() {
    return (
      <div className="fixed right-4 top-4 z-50 rounded-md bg-red-500 px-4 py-2 text-white transition hover:bg-red-600">
        <div className="flex items-center space-x-2">
          <span className="text-3xl">
            <i className="bx bx-check"></i>
          </span>
          <p className="font-bold">{this.title}</p>
          {this.description && <p>this.description</p>}
        </div>
      </div>
    );
  }
}

export default ErrorNotification;
