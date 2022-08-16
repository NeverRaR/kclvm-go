import typing
import json
import inspect
import sys

import kclvm.kcl.info as kcl_info
import kclvm.compiler.extension.plugin.plugin as kcl_plugin

_plugin_dict = {}

def _call_py_method(name: str, args_json: str, kwargs_json: str) -> str:
        try:
            return _call_py_method_unsafe(name, args_json, kwargs_json)
        except Exception as e:
            return json.dumps({"__kcl_PanicInfo__": f"{e}"})

def _call_py_method_unsafe(
        name: str, args_json: str, kwargs_json: str
    ) -> str:
        dotIdx = name.rfind(".")
        if dotIdx < 0:
            return ""
        
        modulePath = name[:dotIdx]
        mathodName = name[dotIdx + 1 :]

        plugin_name = modulePath[modulePath.rfind(".") + 1 :]

        module = _get_plugin(plugin_name)
        mathodFunc = None
        for func_name, func in inspect.getmembers(module):
            if func_name == kcl_info.demangle(mathodName):
                mathodFunc = func
                break
        args = []
        kwargs = {}

        if args_json:
            args = json.loads(args_json)
            if not isinstance(args, list):
                return ""
        if kwargs_json:
            kwargs = json.loads(kwargs_json)
            if not isinstance(kwargs, dict):
                return ""

        result = mathodFunc(*args, **kwargs)
        sys.stdout.flush()
        return json.dumps(result)

def _get_plugin(plugin_name: str) -> typing.Optional[any]:
    if plugin_name in _plugin_dict:
        return _plugin_dict[plugin_name]
    
    module = kcl_plugin.get_plugin(plugin_name)
    _plugin_dict[plugin_name] = module
    return module

def hello(name :str) -> str:
    return "hello plugin : " +name