import React from 'react';

const UserNamesDropdown = ({ users }) => {
    if (!users) {
        return null;
    }

    return (
        <div className="
            p-6
            rounded-lg
            shadow-xl
            max-w-sm mx-auto
        ">
            <div className="flex flex-col">
                <label htmlFor="userNameSelect" className="
                    block
                    text-gray-300
                    text-ml font-bold
                    mb-4
                ">
                    Choose your name:
                </label>
                <select
                    id="userNameSelect"
                    name="userNameSelect"
                    className="
                        bg-gray-800
                        block w-full px-4 py-2
                        rounded-md shadow-inner
                        bg-black
                        text-red-600
                        focus:outline-none
                        ml:text-ml
                        appearance-none
                        pr-8
                    "
                >
                    <option value="" className="bg-gray-800 text-red-300">Select your name...</option>
                    {users.map(user => (
                        <option
                            key={user.name}
                            value={user.name}
                            className="bg-gray-800 text-red-300 hover:bg-red-900"
                        >
                            {user.name}
                        </option>
                    ))}
                </select>
            </div>
        </div>
    );
};

export default UserNamesDropdown;