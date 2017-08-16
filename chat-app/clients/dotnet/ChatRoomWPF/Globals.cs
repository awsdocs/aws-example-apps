using System;
using System.Collections.Generic;
using System.Linq;
using System.Web;

namespace PersistedVariables
{
    public static class P
    {
        // parameterless constructor required for static class
        static P() { StartRow = -1; } // default value

        // public get, and private set for strict access control
        public static int StartRow { get; set; }
        public static int RowCount { get; set; }
        public static string AccessToken { get; set; }
        public static string RefreshToken { get; set; }
    }
}