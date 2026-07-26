import type { Metadata } from "next";
import { GamesHall } from "./games-hall";
import "./styles.css";

export const metadata:Metadata={title:"Games | Obiara"};
export default function GamesPage(){return <GamesHall/>}
