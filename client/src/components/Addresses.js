import React, { useState } from "react";
import axios from "axios";
import { Button, Dialog, TextField, Typography } from "@mui/material";
import { Box } from "@mui/system";
import { useNavigate } from "react-router-dom";

const Addresses = () => {
  const [list, setList] = useState([]);
  const [open, setOpen] = useState(false);
  const [address, setAddress] = useState("");
  const [openC, setOpenC] = useState(false);
  const [addressC, setAddressC] = useState("");
  let navigate = useNavigate();
  React.useEffect(() => {
    (async () => {
      const res = await axios.get("http://localhost:8080/listaddresses");
      setList(res.data.data || []);
    })();
  }, []);

  return (
    <div className="df fdc">
      <Box className="df jcsb">
        <Typography className="sb df" variant="h6">
          List Addresses In Chain
        </Typography>
        <Box className="df">
          <Button
            variant="outlined"
            sx={{ mr: 3 }}
            onClick={async () => {
              navigate("/history");
            }}
          >
            Chain History
          </Button>
          <Button
            variant="outlined"
            sx={{ mr: 3 }}
            onClick={async () => {
              const res = await axios.post(
                "http://localhost:8080/createwallet"
              );
              setList((list) => [...list, res.data.data]);
              setAddressC(res.data.data);
              setOpenC(true);
            }}
          >
            Create Wallet
          </Button>
          <Button
            variant="contained"
            onClick={() => {
              setOpen(true);
            }}
          >
            Import Wallet
          </Button>
        </Box>
      </Box>
      {list.map((item, index) => {
        return (
          <Box
            key={index}
            sx={{
              p: 2,
              m: 2,
              borderRadius: 2,
              border: "1px solid #e8e8e8",
              cursor: "pointer",
              transition: "0.25s ease all",

              "&:hover": {
                background: "#1e88e5",
                color: "#fff",
              },
            }}
            onClick={() => {
              navigate(`/${item}`);
            }}
          >
            {index + 1}. {item}
          </Box>
        );
      })}
      {open && (
        <Dialog
          open={open}
          onClose={() => {
            setOpen(false);
            setAddress("");
          }}
          maxWidth="sm"
          fullWidth={true}
        >
          <Box sx={{ p: 3 }}>
            <Typography className="sb" sx={{ mb: 2 }}>
              Import your wallet
            </Typography>
            <TextField
              label="Address"
              fullWidth
              value={address}
              onChange={(e) => {
                setAddress(e.target.value);
              }}
            />
            <Box className="df" sx={{ mt: 2, justifyContent: "flex-end" }}>
              <Button
                variant="outlined"
                sx={{ mr: 3 }}
                onClick={() => {
                  setOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                variant="contained"
                onClick={async () => {
                  const res = await axios.get(
                    "http://localhost:8080/getbalance",
                    {
                      params: {
                        address,
                      },
                    }
                  );

                  if (typeof res.data.balance == "number") {
                    navigate(`${address}`);
                  }
                }}
              >
                Import
              </Button>
            </Box>
          </Box>
        </Dialog>
      )}
      {openC && (
        <Dialog
          open={openC}
          onClose={() => {
            setOpenC(false);
            setAddressC("");
          }}
          maxWidth="sm"
          fullWidth={true}
        >
          <Box sx={{ p: 2 }}>
            <Typography className="sb" sx={{ mb: 2 }}>
              New Address
            </Typography>
            <Typography>{addressC}</Typography>
          </Box>
        </Dialog>
      )}
    </div>
  );
};

export default Addresses;
