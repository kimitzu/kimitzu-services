// Symbol Polyfill

var Symbol = function(){};
Array.prototype[Symbol.iterator = "@@iterator"] = function()
{
    var i = -1;
    var arr = this;
    return {
        next: function()
        {
            i++;
            return {
                value: arr[i],
                done: i >= arr.length
            };
        },
        return: function()
        {
            return true;
        }
    };
}