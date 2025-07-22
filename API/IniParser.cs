namespace API;

public class IniParser
{
    private static Dictionary<string, Dictionary<string, string>> _iniData = [];
    public static Dictionary<string, Dictionary<string, string>> ParseIniFile(string filePath)
    {
        if(_iniData.Count > 0)
        {
            return _iniData; // Return cached data if already parsed
        }

        if (string.IsNullOrWhiteSpace(filePath))
        {
            throw new ArgumentException("File path cannot be null or empty.", nameof(filePath));
        }

        if (!File.Exists(filePath))
        {
            throw new FileNotFoundException($"The specified INI file does not exist: {filePath}");
        }

        var iniData = new Dictionary<string, Dictionary<string, string>>();
        string? currentSection = null;
        foreach (var line in File.ReadLines(filePath))
        {
            var trimmedLine = line.Trim();
            if (string.IsNullOrEmpty(trimmedLine) || trimmedLine.StartsWith(";"))
            {
                continue; // Skip empty lines and comments
            }
            if (trimmedLine.StartsWith("[") && trimmedLine.EndsWith("]"))
            {
                currentSection = trimmedLine[1..^1].Trim(); // Extract section name
                iniData[currentSection] = new Dictionary<string, string>();
            }
            else if (currentSection != null)
            {
                var keyValue = trimmedLine.Split('=', 2);
                if (keyValue.Length == 2)
                {
                    var key = keyValue[0].Trim();
                    var value = keyValue[1].Trim();
                    iniData[currentSection][key] = value;
                }
            }
        }
        _iniData = iniData;
        return iniData;
    }
}
