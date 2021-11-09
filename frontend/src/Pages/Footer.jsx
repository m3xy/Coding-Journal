import React from "react";

function Footer() {
	return (
		<div className="footer" style={{paddingBottom: '150px', minHeight: '100vh', overflow:'hidden', position:'relative', display:'block'}}>
			<footer className="py-5 bg-dark fixed-bottom" style={{position:'absolute', bottom:0, width:'100%'}}>
				<div className="container">
					<p className="m-0 text-center text-white">
						Copyright &copy; Your Website 2021
					</p>
				</div>
			</footer>
		</div>
	)
}

export default Footer;
