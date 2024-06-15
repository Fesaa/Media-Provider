import React, {useState} from "react";
import NotificationHandler from "../notifications/handler";
import { TrashIcon } from "@heroicons/react/16/solid";
import {InfoStat, SpeedData, StopRequest} from "../utils/types";
import {stopDownload} from "../utils/http";
import {ChevronDownIcon, ChevronUpIcon} from "@heroicons/react/24/outline";
import {Line} from "react-chartjs-2";
import {ChartData, ChartOptions} from "chart.js";
import DirBrowser from "./io/browser";

function typeConv(type: string): string {
    switch (type) {
      case "bytes":
        return "MB/s";
      case "volumes":
        return "volumes/s"
      case "images":
        return "images/s"
    }
    return type
}

function speedConv(speed:number, type: string): number {
  switch (type) {
    case "bytes":
      return +(speed / (1024 * 1024)).toFixed(2);
  }
  return speed
}

export default function InfoLine(props: {
  TKey: string;
  infoStat: InfoStat;
  speeds: SpeedData[];
  refreshFunc: (repeat: boolean) => void;
}) {
  const [open, setOpen] = useState(false);
  const i = props.infoStat;
  const toggleOpen = () => setOpen(!open);

  async function remove(id: string) {
    const stopRequest: StopRequest = {
      provider: i.provider,
      id: id,
      delete_files: true,
    }
    stopDownload(stopRequest).then(() => {
      NotificationHandler.addSuccesNotificationByTitle("Successfully stop content download");
    }).catch((err) => {
        NotificationHandler.addErrorNotificationByTitle(err.message);
    })
  }

  const chartOptions: ChartOptions<"line"> = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      y: {
        beginAtZero: true,
      }
    },
    plugins: {
      title: {
        display: true,
        text: `Download speed (${speedConv(i.speed.speed, i.speed_type)} ${typeConv(i.speed_type)})`,
      }
    }
  }

  const chartData: ChartData<"line", number[]> = {
    labels: props.speeds.map((s) => (new Date(s.time)).toLocaleTimeString()),
    datasets: [{
      label: 'Download speed',
      data: props.speeds.map((s) => s.speed).map((s) => speedConv(s, i.speed_type)),
      fill: false,
      borderColor: 'rgb(75, 192, 192)',
      tension: 0.1,
    }]
  }

  return (
    <div className="flex flex-grow flex-col bg-white border-2 border-solid border-gray-200 p-5 text-center" key={i.id}>
      <div className={`space-x-2 flex flex-row ${open && "pb-2 border-b-2 border-gray-200 border-solid"}`}>

        <div className="flex flex-col flex-grow">
          <div className="flex flex-row justify-between">
            <div className={`space-x-2 flex flex-row`} onClick={() => toggleOpen()}>
              {open
                  ? <ChevronUpIcon className="hidden md:block h-6 w-6"/>
                  : <ChevronDownIcon className="hidden md:block h-6 w-6"/>
              }
              <span className="break-all min-w-20 md:min-w-56">
                  {i.name}
                </span>
            </div>
            <span className="break-all min-w-20 md:min-w-56 flex flex-row space-x-4">
              <span>{i.progress} %</span>
              <span className="hidden md:block">
                @ {speedConv(i.speed.speed, i.speed_type)} {typeConv(i.speed_type)}
              </span>
              <span>
                ({i.size})
              </span>
            </span>
          </div>

          <div className="flex flex-col flex-grow justify-center mt-3">
            <div className="h-2.5 rounded-full w-full bg-gray-200 dark:bg-gray-700  md:block">
              <div
                  className="h-2.5 rounded-full bg-blue-600"
                  style={{width: `${i.progress}%`}}
              ></div>
            </div>
          </div>
        </div>
        <div className="flex flex-col justify-center">
          <TrashIcon className="h-8 md:h-12 w-8 md:w-12 text-red-500 hover:cursor-pointer"
                     onClick={() => remove(i.id)}/>
          {open
              ? <ChevronUpIcon className="md:hidden h-6 w-6" onClick={() => toggleOpen()} />
              : <ChevronDownIcon className="md:hidden h-6 w-6" onClick={() => toggleOpen()} />
          }
        </div>
      </div>
      {open && <div className="flex flex-col md:flex-row justify-around">
        <div className="w-full max-w-md h-96">
          <Line options={chartOptions} data={chartData}/>
        </div>
        {i.download_dir && <div className="flex flex-col flex-grow mx-5 justify-center">
          <DirBrowser base={i.download_dir} name={i.name} addFiles={false} showFiles={true} copy={false} />
        </div>}
      </div>}
    </div>
  );
}
