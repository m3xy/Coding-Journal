import React, { useState, useEffect } from "react"
import styles from "./HomePage.module.css"
import SubmissionList from "../../Components/SubmissionList"
import JwtService from "../../Web/jwt.service"

export default () => {
	const [list, setList] = useState([
		{
			display: "Assigned submissions",
			condition: () => {
				const userType = JwtService.getUserType()
				return userType == 2 || userType == 3
			},
			query: {
				reviewers: [JwtService.getUserID()]
			}
		},
		{
			display: "Most recent submissions",
			condition: () => {
				return true
			},
			query: {
				orderBy: "newest"
			}
		}
	])

	return (
		<div className={styles.HomePage}>
			<div className={styles.HomeContent}>
				{list.map((recipe, i) => {
					if (recipe.condition()) {
						return (
							<div key={i} style={{ marginBottom: "100px" }}>
								<SubmissionList {...recipe} />
							</div>
						)
					} else {
						return <div key={i}></div>
					}
				})}
			</div>
		</div>
	)
}
