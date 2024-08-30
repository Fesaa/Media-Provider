import React from "react";
import {Modifier} from "./types";
import {ChevronDownIcon, ChevronUpIcon} from "@heroicons/react/24/outline";

type onChangeFunc = (modifierKey: string, newValue: string | string[]) => void;

export interface ModifierProps {
    keyProp: string;
    modifier: Modifier;
    onChange: onChangeFunc;
}

class DropDownModifier extends React.Component<ModifierProps> {
    constructor(props: ModifierProps) {
        super(props);
    }

    render() {
        return <div key={this.props.keyProp}>
            <label
                htmlFor={this.props.keyProp}
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
            >
                {this.props.modifier.title}
            </label>
                <select
                    onChange={(e) => this.props.onChange(this.props.keyProp, e.target.value)}
                    className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-gray-900 sm:text-sm"
                >

                    {Object.entries(this.props.modifier.values).map(([key, value]) => {
                        return <option key={key} value={key}>
                            {value}
                        </option>
                    })}
                </select>
        </div>;
    }
}

interface MultiSelectState {
    checkedItems: string[];
    showOptions: boolean;
    query: string;
}

class MultiSelectModifier extends React.Component<ModifierProps, MultiSelectState> {
    constructor(props: ModifierProps) {
        super(props);
        this.state = {
            showOptions: false,
            checkedItems: [],
            query: "",
        }

        this.toggleOptions.bind(this)
        this.handleCheckboxChange.bind(this)
    }

    private toggleOptions() {
        this.setState({...this.state, showOptions: !this.state.showOptions});
    }

    private handleCheckboxChange(value: string) {
        const newCheckedItems = this.state.checkedItems.includes(value)
            ? this.state.checkedItems.filter(item => item !== value)
            : [...this.state.checkedItems, value];
        this.setState({...this.state, checkedItems: newCheckedItems});
        this.props.onChange(this.props.keyProp, newCheckedItems);
    }

    private normalize(value: string): string {
        return value.toLowerCase().replace(/[^a-z0-9]/g, '');
    }

    render() {
        return <div>
            <div className="flex items-center justify-between hover:cursor-pointer" onClick={() => this.toggleOptions()}>
                <h2 className="text-lg font-semibold">{this.props.modifier.title} ({this.state.checkedItems.length}) </h2>
                <div className="px-4 py-2 rounded">
                    {this.state.showOptions ? <ChevronUpIcon
                        className="-mr-1 h-5 w-5 text-gray-400"
                        aria-hidden="true"
                    /> : <ChevronDownIcon
                        className="-mr-1 h-5 w-5 text-gray-400"
                        aria-hidden="true"
                    />}
                </div>
            </div>
            {this.state.showOptions && (
                <div className="flex flex-col space-y-2">
                    {<input
                        className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-1 text-gray-900 sm:text-sm"
                        type="text"
                        placeholder="Filter..."
                        onChange={(e) => {
                            this.setState({...this.state, query: e.target.value})
                    }} />}
                    <div className="grid grid-cols-3 gap-4 mt-4 max-h-48 overflow-auto">
                        {Object.entries(this.props.modifier.values)
                            .filter(option => this.normalize(option[0]).includes(this.normalize(this.state.query)))
                            .map(option => {
                                const [key, value] = option as [string, string];
                                return <div key={key} className="flex items-center hover:cursor-pointer">
                                    <input
                                        type="checkbox"
                                        id={key}
                                        value={value}
                                        checked={this.state.checkedItems.includes(key)}
                                        onChange={(e) => this.handleCheckboxChange(key)}
                                        className="rounded border-gray-300 text-primary-600 focus:ring-primary-600 focus:border-primary-600 shadow-sm focus:ring focus:ring-opacity-50 h-4 w-4 mr-2"
                                    />
                                    <label htmlFor={key}
                                           className="text-xs hover:cursor-pointer">{value}</label>
                                </div>
                            })}
                    </div>
                </div>
            )}
        </div>
    }
}

export function GetModifierComponent(key: string, modifier: Modifier, onChange: onChangeFunc): React.ReactNode {
    switch (modifier.type) {
        case "dropdown":
            return <DropDownModifier keyProp={key} modifier={modifier} onChange={onChange}/>;
        case "multi":
            return <MultiSelectModifier keyProp={key} modifier={modifier} onChange={onChange}/>;
        default:
            return <div></div>
    }
}