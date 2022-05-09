import logo from "./logo.svg";
import "./App.css";
import Addresses from "./components/Addresses";
import { Container } from "@mui/material";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Wallet from "./components/Wallet";
import History from "./components/History";

function App() {
  return (
    <Container
      className="App df"
      sx={{
        padding: 4,
      }}
      maxWidth="md"
    >
      <BrowserRouter>
        <Routes>
          <Route path="/" exact element={<Addresses />} />
          <Route path=":address" element={<Wallet />}></Route>
          <Route path="/history" element={<History />} />
        </Routes>
      </BrowserRouter>
    </Container>
  );
}

export default App;
