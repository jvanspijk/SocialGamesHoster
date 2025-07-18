import { useNavigate } from "react-router";

function App() {    
    const navigate = useNavigate();

    return (
        <div style={{ display: "flex", flexDirection: "column", alignItems: "center", marginTop: "100px" }}>
            <button onClick={() => navigate("/game")} style={{ margin: "10px", padding: "10px 20px" }}>
                For players
            </button>
            <button onClick={() => navigate("/admin")} style={{ margin: "10px", padding: "10px 20px" }}>
                For admins
            </button>
        </div>
    );
}

export default App;