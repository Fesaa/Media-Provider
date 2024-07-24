import { GlobeAltIcon } from "@heroicons/react/16/solid";
import React, { Dispatch, SetStateAction, useState } from "react";
import DirBrowser from "./browser";

type DirFormComponentProps = {
    base: string;
    name: string;
    setter: Dispatch<SetStateAction<string>>;
}

export default function DirFormComponent(props: DirFormComponentProps) {
    const [showPopup, setShowPopup] = useState<boolean>(false);
    const [inputValue, setInputValue] = useState<string>("");

    return <div>
        <label
            htmlFor="custom_dir"
            className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
        >
            Custom Directory
        </label>
        <div className="flex flex-row items-center text-center space-x-2">
            <input
                type="text"
                name="custom_dir"
                id="custom_dir"
                className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-gray-900 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
                required
                value={inputValue}
                onChange={(e) => {
                    props.setter(e.target.value)
                    setInputValue(e.target.value)
                }}
            />
            <GlobeAltIcon className="w-8 h-8 hover:cursor-pointer" onClick={_ => setShowPopup(!showPopup)} />
        </div>
        {showPopup && <DirBrowser
            base={props.base}
            name={props.name}
            showFiles={false}
            addFiles={true}
            callback={(path) => setInputValue(path)}
        />}
    </div>
}

