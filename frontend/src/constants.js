/**
 * constants.js
 * author: 190010425
 * created: October 27, 2021
 * 
 * This file holds constants for the project. This file exists
 * to prevent inconsistency in things like host name or host port
 * and the bugs which come with said inconsistency
 * 
 * Note that the values of the constants can be changed without changing
 * the other code files, but if the structure of the constants objects
 * are changed, it must be accompanied by similar changes to the other
 * code files
 */


/**
 * Information on this react app, such as where it is hosted
 */
const FRONTEND_INFO = {
    host: "http://localhost", // TEMP: for now
    port: 3000, // TEMP: also for now
}

/**
 * Information for communicating with the backend from this react app
 */
const BACKEND_INFO = {
    host: "http://localhost",
    port: 3333,
}

export { BACKEND_INFO, FRONTEND_INFO };