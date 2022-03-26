/*
 * LoggedInDropdown.jsx
 * Author: 190014935
 *
 * Button and dropdown menu for logged in users.
 */
import React, { useEffect, useState } from "react"
import { Dropdown } from "react-bootstrap"
import { useNavigate } from "react-router-dom"
import JwtService from "../../../Web/jwt.service"

const LoggedInDropdown = ({ user }) => {
	const navigate = useNavigate()

	const permissionLvls = [
		"user",
		"author",
		"reviewer",
		"author-reviewer",
		"editor"
	]
	const [perm, setPermissions] = useState(0)

	useEffect(() => {
		setPermissions(JwtService.getUserType())
	})

	return (
		<Dropdown>
			<Dropdown.Toggle
				variant={
					permissionLvls[perm] === "editor" ? "warning" : "primary"
				}>
				{user.firstName + " " + user.lastName}
			</Dropdown.Toggle>
			<Dropdown.Menu variant="dark" align="end">
				<Dropdown.Item
					onClick={() => {
						navigate("/profile")
					}}>
					{" "}
					Profile{" "}
				</Dropdown.Item>
				<Dropdown.Item
					onClick={() => {
						navigate("/submissions")
					}}>
					{" "}
					Submissions{" "}
				</Dropdown.Item>
				<Dropdown.Item
					onClick={() => {
						JwtService.rmUser()
						navigate("/")
					}}>
					{" "}
					Log Out{" "}
				</Dropdown.Item>
			</Dropdown.Menu>
		</Dropdown>
	)
}
export default LoggedInDropdown
