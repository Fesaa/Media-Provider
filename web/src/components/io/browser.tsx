import React from "react";
import Dir from "./dir";

export default function DirBrowser(props: { base: string, name: string }) {
    return <div className="mt-4 focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 text-gray-900 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500 p-2.5">
        <span className="text-xl mb-5">Directory browser</span>
        <div className="ml-2">
            <Dir base={props.base} depth={1} name={props.name} />
        </div>
    </div>
}