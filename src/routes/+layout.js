export const prerender = true;
export const ssr = false;

import "bulma/css/bulma.css";
import "../app.css";
import dayjs from "dayjs";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
dayjs.extend(isSameOrBefore);