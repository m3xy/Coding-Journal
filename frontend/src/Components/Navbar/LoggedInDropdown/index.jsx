/*
 * LoggedInDropdown.jsx
 * Author: 190014935
 *
 * Button and dropdown menu for logged in users.
 */
import React from "react"
import { Dropdown } from "react-bootstrap"
import { useNavigate } from "react-router-dom"

const LoggedInDropdown = ({user}) => {
	const navigate = useNavigate()
	return(
		<Dropdown>
			<Dropdown.Toggle>
				{user.firstName + " " + user.lastName}
			</Dropdown.Toggle>
			<Dropdown.Menu variant="dark" align="end">
				<Dropdown.Item
					onClick={() => { navigate("/profile") }}>
					{" "}
					Profile{" "}
				</Dropdown.Item>
				<Dropdown.Item
					onClick={() => { navigate("/submissions") }}>
					{" "}
					Submissions{" "}
				</Dropdown.Item>
				<Dropdown.Item
					onClick={() => { JwtService.rmUser(); navigate("/")}}
				>
					{" "}
					Log Out{" "}	
				</Dropdown.Item>
			</Dropdown.Menu>
		</Dropdown>
	)
}
export default LoggedInDropdown
